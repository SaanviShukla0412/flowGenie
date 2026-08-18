package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/SaanviShukla0412/flowGenie/internal/database"
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
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/workflows/")
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

		if err := redisQueue.EnqueueWorkflow(r.Context(), *wf); err != nil {
			http.Error(w, "failed to enqueue workflow", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		json.NewEncoder(w).Encode(map[string]string{
			"status":      "queued",
			"workflow_id": wf.ID,
		})
	})

	log.Println("FlowGenie server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
