package queue

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

type WorkflowJob struct {
	ExecutionID string            `json:"execution_id"`
	Workflow    workflow.Workflow `json:"workflow"`
}

const WorkflowQueue = "workflow_jobs"

type RedisQueue struct {
	Client *redis.Client
}

func NewRedisQueue() *RedisQueue {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return &RedisQueue{
		Client: client,
	}
}

func (q *RedisQueue) Ping(ctx context.Context) error {
	return q.Client.Ping(ctx).Err()
}

func (q *RedisQueue) EnqueueWorkflow(
	ctx context.Context,
	executionID string,
	wf workflow.Workflow,
) error {
	job := WorkflowJob{
		ExecutionID: executionID,
		Workflow:    wf,
	}

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return q.Client.RPush(ctx, WorkflowQueue, data).Err()
}

func (q *RedisQueue) DequeueWorkflow(
	ctx context.Context,
) (WorkflowJob, error) {
	result, err := q.Client.BLPop(
		ctx,
		0,
		WorkflowQueue,
	).Result()

	if err != nil {
		return WorkflowJob{}, err
	}

	var job WorkflowJob

	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return WorkflowJob{}, err
	}

	return job, nil
}
