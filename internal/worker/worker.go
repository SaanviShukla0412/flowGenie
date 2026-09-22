package worker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/SaanviShukla0412/flowGenie/internal/database"
	"github.com/SaanviShukla0412/flowGenie/internal/execution"
	"github.com/SaanviShukla0412/flowGenie/internal/executor"
	"github.com/SaanviShukla0412/flowGenie/internal/queue"
	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

type Worker struct {
	Queue    *queue.RedisQueue
	Executor *executor.Executor
	DB       *pgx.Conn
}

func NewWorker(
	q *queue.RedisQueue,
	e *executor.Executor,
	db *pgx.Conn,
) *Worker {
	return &Worker{
		Queue:    q,
		Executor: e,
		DB:       db,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started")

	for {
		job, err := w.Queue.DequeueWorkflow(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Worker shutting down")
				return
			}

			log.Println("Failed to dequeue workflow:", err)
			continue
		}

		wf := job.Workflow

		log.Printf(
			"Received execution: id=%s workflow_id=%s",
			job.ExecutionID,
			wf.ID,
		)

		if err := database.UpdateExecutionStatus(
			w.DB,
			job.ExecutionID,
			"running",
		); err != nil {
			log.Printf(
				"Failed to update execution status: id=%s error=%v",
				job.ExecutionID,
				err,
			)
			continue
		}

		executionContext := execution.NewContext()
		workflowFailed := false

		for _, step := range wf.Steps {
			executionStepID := "step_" + uuid.New().String()
			if err := database.CreateExecutionStep(
				w.DB,
				executionStepID,
				job.ExecutionID,
				step.Name,
				step.Type,
				"running",
			); err != nil {
				log.Printf(
					"Failed to create execution step: name=%s error=%v",
					step.Name,
					err,
				)
			}
			log.Printf(
				"Executing step: name=%s type=%s",
				step.Name,
				step.Type,
			)

			resolvedStep := resolveStepTemplates(step, executionContext)

			output, err := w.Executor.Execute(ctx, resolvedStep)
			if err != nil {
				errorMessage := err.Error()

				if updateErr := database.UpdateExecutionStep(
					w.DB,
					executionStepID,
					"failed",
					nil,
					&errorMessage,
				); updateErr != nil {
					log.Printf(
						"Failed to update execution step: name=%s error=%v",
						step.Name,
						updateErr,
					)
				}

				log.Printf(
					"Step failed: name=%s error=%v",
					step.Name,
					err,
				)

				workflowFailed = true
				break
			}

			log.Printf(
				"Step completed: name=%s output=%s",
				step.Name,
				output,
			)
			if updateErr := database.UpdateExecutionStep(
				w.DB,
				executionStepID,
				"completed",
				&output,
				nil,
			); updateErr != nil {
				log.Printf(
					"Failed to update execution step: name=%s error=%v",
					step.Name,
					updateErr,
				)
			}
			parsedOutput := parseStepOutput(output)

			executionContext.SetOutput(step.Name, parsedOutput)
		}

		if workflowFailed {
			if err := database.UpdateExecutionStatus(
				w.DB,
				job.ExecutionID,
				"failed",
			); err != nil {
				log.Printf(
					"Failed to update execution status: id=%s error=%v",
					job.ExecutionID,
					err,
				)
				continue
			}

			log.Printf(
				"Workflow failed: id=%s",
				wf.ID,
			)
			continue
		}

		if err := database.UpdateExecutionStatus(
			w.DB,
			job.ExecutionID,
			"completed",
		); err != nil {
			log.Printf(
				"Failed to update execution status: id=%s error=%v",
				job.ExecutionID,
				err,
			)
			continue
		}

		log.Printf(
			"Workflow completed: id=%s",
			wf.ID,
		)
	}
}

func resolveStepTemplates(
	step workflow.Step,
	executionContext *execution.Context,
) workflow.Step {
	resolvedStep := step

	if step.Config == nil {
		return resolvedStep
	}

	resolvedConfig := make(map[string]interface{})

	for key, value := range step.Config {
		stringValue, ok := value.(string)
		if !ok {
			resolvedConfig[key] = value
			continue
		}

		resolvedConfig[key] = execution.ResolveTemplate(
			stringValue,
			executionContext,
		)
	}

	resolvedStep.Config = resolvedConfig

	return resolvedStep
}

func parseStepOutput(output string) interface{} {
	var parsed map[string]interface{}

	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return output
	}

	return parsed
}
