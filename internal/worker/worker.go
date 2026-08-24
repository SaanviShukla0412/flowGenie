package worker

import (
	"context"
	"log"

	"github.com/SaanviShukla0412/flowGenie/internal/executor"
	"github.com/SaanviShukla0412/flowGenie/internal/queue"
)

type Worker struct {
	Queue    *queue.RedisQueue
	Executor *executor.Executor
}

func NewWorker(
	q *queue.RedisQueue,
	e *executor.Executor,
) *Worker {
	return &Worker{
		Queue:    q,
		Executor: e,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started")

	for {
		wf, err := w.Queue.DequeueWorkflow(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Worker shutting down")
				return
			}

			log.Println("Failed to dequeue workflow:", err)
			continue
		}

		log.Printf(
			"Received workflow: id=%s name=%s",
			wf.ID,
			wf.Name,
		)

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
			log.Printf(
				"Workflow failed: id=%s",
				wf.ID,
			)
			continue
		}

		log.Printf(
			"Workflow completed: id=%s",
			wf.ID,
		)
	}
}
