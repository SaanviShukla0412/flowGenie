package queue

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

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
	wf workflow.Workflow,
) error {
	data, err := json.Marshal(wf)
	if err != nil {
		return err
	}
	return q.Client.RPush( // RPush : pushing somethin to the rightmost or in the last of the redis list
		ctx,
		WorkflowQueue,
		data,
	).Err()
}
