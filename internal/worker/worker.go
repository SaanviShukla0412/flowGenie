package worker

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/SaanviShukla0412/flowGenie/internal/database"
	"github.com/SaanviShukla0412/flowGenie/internal/executor"
	"github.com/SaanviShukla0412/flowGenie/internal/queue"
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

		workflowFailed := false

		for _, step := range wf.Steps {
			log.Printf(
				"Executing step: name=%s type=%s",
				step.Name,
				step.Type,
			)

			if err := w.Executor.Execute(ctx, step); err != nil {
				log.Printf(
					"Step failed: name=%s error=%v",
					step.Name,
					err,
				)

				workflowFailed = true
				break
			}
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
