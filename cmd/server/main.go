package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/SaanviShukla0412/flowGenie/internal/database"
	"github.com/SaanviShukla0412/flowGenie/internal/execution"
	"github.com/SaanviShukla0412/flowGenie/internal/queue"
	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

func main() {
	// Connect to PostgreSQL
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close(context.Background())
	log.Println("Connected to PostgreSQL")

	// Connect to Redis
	redisQueue := queue.NewRedisQueue()
	if err := redisQueue.Ping(context.Background()); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Connected to Redis")

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	// Create workflow
	http.HandleFunc("/workflows", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var wf workflow.Workflow
		if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := workflow.Validate(wf); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		wf.ID = "wf_" + uuid.New().String()
		stepsJSON, err := json.Marshal(wf.Steps)
		if err != nil {
			http.Error(w, "failed to encode workflow steps", http.StatusInternalServerError)
			return
		}
		if err := database.CreateWorkflow(
			db,
			wf.ID,
			wf.Name,
			stepsJSON,
		); err != nil {
			http.Error(w, "failed to create workflow", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(wf)
	})

	// Run workflow
	http.HandleFunc("/workflows/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/workflows/")

		// Get execution history for a workflow
		if strings.HasSuffix(path, "/executions") {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			workflowID := strings.TrimSuffix(path, "/executions")

			if workflowID == "" {
				http.Error(w, "invalid workflow path", http.StatusBadRequest)
				return
			}

			executions, err := database.GetExecutionsByWorkflowID(
				db,
				workflowID,
			)
			if err != nil {
				http.Error(w, "failed to get executions", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(executions)
			return
		}

		// Run workflow
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimSuffix(path, "/run")

		if id == path || id == "" {
			http.Error(w, "invalid workflow path", http.StatusBadRequest)
			return
		}

		wf, err := database.GetWorkflowByID(db, id)
		if err != nil {
			http.Error(w, "workflow not found", http.StatusNotFound)
			return
		}

		executionID := "exec_" + uuid.New().String()

		if err := database.CreateExecution(
			db,
			executionID,
			wf.ID,
			"queued",
		); err != nil {
			http.Error(w, "failed to create execution", http.StatusInternalServerError)
			return
		}

		if err := redisQueue.EnqueueWorkflow(r.Context(), executionID, *wf); err != nil {
			if updateErr := database.UpdateExecutionStatus(
				db,
				executionID,
				"failed",
			); updateErr != nil {
				log.Printf(
					"Failed to update execution status: id=%s error=%v",
					executionID,
					updateErr,
				)
			}

			http.Error(w, "failed to enqueue workflow", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		json.NewEncoder(w).Encode(map[string]string{
			"execution_id": executionID,
			"workflow_id":  wf.ID,
			"status":       "queued",
		})
	})

	http.HandleFunc("/executions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		executionID := strings.TrimPrefix(r.URL.Path, "/executions/")

		if executionID == "" {
			http.Error(w, "invalid execution path", http.StatusBadRequest)
			return
		}

		exec, err := database.GetExecutionByID(db, executionID)
		if err != nil {
			http.Error(w, "execution not found", http.StatusNotFound)
			return
		}

		steps, err := database.GetExecutionStepsByExecutionID(
			db,
			executionID,
		)
		if err != nil {
			http.Error(w, "failed to get execution steps", http.StatusInternalServerError)
			return
		}

		response := struct {
			Execution *execution.Execution      `json:"execution"`
			Steps     []execution.ExecutionStep `json:"steps"`
		}{
			Execution: exec,
			Steps:     steps,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr: ":8080",
	}

	go func() {
		log.Println("FlowGenie server running on :8080")

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("Shutting down FlowGenie server...")

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	log.Println("FlowGenie server stopped")
}
