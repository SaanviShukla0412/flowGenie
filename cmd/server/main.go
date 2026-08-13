package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/SaanviShukla0412/flowGenie/internal/database"
	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status":"ok"}`))
}

func createWorkflowHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var wf workflow.Workflow
		err := json.NewDecoder(r.Body).Decode(&wf)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		wf.ID = "wf_" + uuid.New().String()
		stepsJSON, err := json.Marshal(wf.Steps)
		if err != nil {
			http.Error(w, "failed to encode steps", http.StatusInternalServerError)
			return
		}

		err = database.CreateWorkflow(
			conn,
			wf.ID,
			wf.Name,
			stepsJSON,
		)
		if err != nil {
			log.Println("Failed to create workflow:", err)
			http.Error(w, "failed to create workflow", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(wf)
	}
}

func getWorkflowsHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		workflows, err := database.GetWorkflows(conn)
		if err != nil {
			log.Println("Failed to get workflows:", err)
			http.Error(w, "failed to get workflows", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(workflows)
	}
}

func getWorkflowByIDHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/workflows/")
		wf, err := database.GetWorkflowByID(conn, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "workflow not found", http.StatusNotFound)
				return
			}
			log.Println("Failed to get workflow:", err)
			http.Error(w, "failed to get workflow", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wf)
	}
}

func main() {

	conn, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(context.Background())
	log.Println("Connected to PostgreSQL")

	http.HandleFunc("/health", healthHandler)
	http.Handle("/workflows", createWorkflowHandler(conn))
	http.Handle("/workflows/list", getWorkflowsHandler(conn))
	http.Handle("/workflows/", getWorkflowByIDHandler(conn))

	log.Println("FlowGenie server running on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
