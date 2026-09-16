package executor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(
	ctx context.Context,
	step workflow.Step,
) (string, error) {
	switch step.Type {
	case "log":
		log.Printf("LOG step executed: %s", step.Name)
		return "", nil

	case "http":
		return e.executeHTTP(ctx, step)

	default:
		return "", fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (e *Executor) executeHTTP(
	ctx context.Context,
	step workflow.Step,
) (string, error) {
	methodValue, ok := step.Config["method"]
	if !ok {
		return "", fmt.Errorf("http step missing method")
	}

	urlValue, ok := step.Config["url"]
	if !ok {
		return "", fmt.Errorf("http step missing url")
	}

	method, ok := methodValue.(string)
	if !ok {
		return "", fmt.Errorf("http step method must be a string")
	}

	url, ok := urlValue.(string)
	if !ok {
		return "", fmt.Errorf("http step url must be a string")
	}

	var bodyReader io.Reader

	if bodyValue, ok := step.Config["body"]; ok {
		body, ok := bodyValue.(string)
		if !ok {
			return "", fmt.Errorf("http step body must be a string")
		}

		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		bodyReader,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	headersValue, ok := step.Config["headers"]
	if ok {
		headers, ok := headersValue.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("http step headers must be an object")
		}

		for key, value := range headers {
			headerValue, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("http step header values must be strings")
			}

			req.Header.Set(key, headerValue)
		}
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read http response: %w", err)
	}

	log.Printf(
		"HTTP step executed: name=%s method=%s url=%s status=%d body=%s",
		step.Name,
		method,
		url,
		response.StatusCode,
		string(body),
	)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf(
			"http request returned status %d",
			response.StatusCode,
		)
	}

	return string(body), nil
}
