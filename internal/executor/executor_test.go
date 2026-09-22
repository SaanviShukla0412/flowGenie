package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestExecuteHTTPTimeout(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"hello"}`))
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Millisecond,
	)
	defer cancel()

	_, err := executor.Execute(
		ctx,
		workflow.Step{
			Name: "Timeout Test",
			Type: "http",
			Config: map[string]interface{}{
				"method": "GET",
				"url":    server.URL,
			},
		},
	)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// ADD THE NEW TEST HERE

func TestExecuteHTTPRetriesOnServerError(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++

			if attempts < 3 {
				http.Error(
					w,
					`{"error":"temporary failure"}`,
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"success"}`))
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	output, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "Retry Test",
			Type: "http",
			Config: map[string]interface{}{
				"method":  "GET",
				"url":     server.URL,
				"retries": float64(2),
			},
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output != `{"message":"success"}` {
		t.Errorf("unexpected output: %s", output)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteHTTPDoesNotRetryOnClientError(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++

			http.Error(
				w,
				`{"error":"bad request"}`,
				http.StatusBadRequest,
			)
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	_, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "No Retry Test",
			Type: "http",
			Config: map[string]interface{}{
				"method":  "GET",
				"url":     server.URL,
				"retries": float64(2),
			},
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestExecuteHTTPPostBodySurvivesRetry(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}

			if string(body) != `{"name":"Saanvi"}` {
				t.Errorf(
					"expected body %s, got %s",
					`{"name":"Saanvi"}`,
					string(body),
				)
			}

			if attempts < 2 {
				http.Error(
					w,
					`{"error":"temporary failure"}`,
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"success"}`))
		}),
	)
	defer server.Close()

	executor := NewExecutor()

	output, err := executor.Execute(
		context.Background(),
		workflow.Step{
			Name: "POST Retry Test",
			Type: "http",
			Config: map[string]interface{}{
				"method":  "POST",
				"url":     server.URL,
				"body":    `{"name":"Saanvi"}`,
				"retries": float64(1),
			},
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output != `{"status":"success"}` {
		t.Errorf("unexpected output: %s", output)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
