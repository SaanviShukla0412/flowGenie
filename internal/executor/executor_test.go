package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

func TestExecuteHTTPGet(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"hello"}`))
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	output, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "Get Test",
			Type: "http",
			Config: map[string]interface{}{
				"method": "GET",
				"url":    server.URL,
			},
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"message":"hello"}`

	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestExecuteHTTPPostWithHeadersAndBody(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf(
					"expected Content-Type application/json, got %s",
					r.Header.Get("Content-Type"),
				)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}

			expectedBody := `{"name":"Saanvi"}`

			if string(body) != expectedBody {
				t.Errorf(
					"expected body %s, got %s",
					expectedBody,
					string(body),
				)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"created"}`))
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	output, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "Create User",
			Type: "http",
			Config: map[string]interface{}{
				"method": "POST",
				"url":    server.URL,
				"headers": map[string]interface{}{
					"Content-Type": "application/json",
				},
				"body": `{"name":"Saanvi"}`,
			},
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"status":"created"}`

	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestExecuteHTTPNon2xxFails(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(
				w,
				`{"error":"not found"}`,
				http.StatusNotFound,
			)
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	_, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "Missing Resource",
			Type: "http",
			Config: map[string]interface{}{
				"method": "GET",
				"url":    server.URL,
			},
		},
	)

	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestExecuteUnknownStepFails(t *testing.T) {
	executor := NewExecutor()

	_, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "Unknown",
			Type: "something_random",
		},
	)

	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
}
