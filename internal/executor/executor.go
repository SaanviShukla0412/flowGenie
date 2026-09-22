package executor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

type Executor struct{}

const (
	defaultHTTPTimeout = 30 * time.Second
	defaultHTTPRetries = 0
)

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultHTTPTimeout,
	}
}

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

	var body string

	if bodyValue, ok := step.Config["body"]; ok {
		bodyValueString, ok := bodyValue.(string)
		if !ok {
			return "", fmt.Errorf("http step body must be a string")
		}

		body = bodyValueString
	}

	headersValue, ok := step.Config["headers"]
	var headers map[string]interface{}

	if ok {
		headers, ok = headersValue.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("http step headers must be an object")
		}
	}

	retries := getRetryCount(step)
	client := newHTTPClient()

	for attempt := 0; attempt <= retries; attempt++ {
		var bodyReader io.Reader

		if body != "" {
			bodyReader = strings.NewReader(body)
		}

		req, err := http.NewRequestWithContext(
			ctx,
			method,
			url,
			bodyReader,
		)
		if err != nil {
			return "", fmt.Errorf(
				"failed to create http request: %w",
				err,
			)
		}

		for key, value := range headers {
			headerValue, ok := value.(string)
			if !ok {
				return "", fmt.Errorf(
					"http step header values must be strings",
				)
			}

			req.Header.Set(key, headerValue)
		}

		response, err := client.Do(req)

		if err != nil {
			if ctx.Err() != nil {
				return "", fmt.Errorf(
					"http request failed: %w",
					ctx.Err(),
				)
			}

			if attempt == retries {
				return "", fmt.Errorf(
					"http request failed after %d attempts: %w",
					attempt+1,
					err,
				)
			}

			log.Printf(
				"HTTP request retrying: name=%s attempt=%d error=%v",
				step.Name,
				attempt+1,
				err,
			)

			continue
		}

		responseBody, err := io.ReadAll(response.Body)
		response.Body.Close()

		if err != nil {
			return "", fmt.Errorf(
				"failed to read http response: %w",
				err,
			)
		}

		log.Printf(
			"HTTP step executed: name=%s method=%s url=%s status=%d body=%s",
			step.Name,
			method,
			url,
			response.StatusCode,
			string(responseBody),
		)

		if response.StatusCode >= 200 &&
			response.StatusCode < 300 {
			return string(responseBody), nil
		}

		if response.StatusCode >= 400 &&
			response.StatusCode < 500 {
			return "",
				fmt.Errorf(
					"http request returned status %d",
					response.StatusCode,
				)
		}

		if attempt == retries {
			return "",
				fmt.Errorf(
					"http request returned status %d after %d attempts",
					response.StatusCode,
					attempt+1,
				)
		}

		log.Printf(
			"HTTP request retrying: name=%s attempt=%d status=%d",
			step.Name,
			attempt+1,
			response.StatusCode,
		)
	}

	return "", fmt.Errorf("http request failed")
}

func getRetryCount(step workflow.Step) int {
	value, ok := step.Config["retries"]
	if !ok {
		return defaultHTTPRetries
	}

	retries, ok := value.(float64)
	if !ok || retries < 0 {
		return defaultHTTPRetries
	}

	return int(retries)
}
