package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kartheek0107/distributed-orchestrator/pkg/models"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	Client *redis.Client
	key    string
}

func NewRedisQueue(addr string, password string, db int) *RedisQueue {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisQueue{
		Client: rdb,
		key:    "task_queue",
	}
}

func (rq *RedisQueue) Enqueue(ctx context.Context, job *models.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	return rq.Client.RPush(ctx, rq.key, data).Err()
}

func (rq *RedisQueue) Dequeue(ctx context.Context) (*models.Job, error) {
	data, err := rq.Client.BLPop(ctx, 0, rq.key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	var job models.Job
	if err := json.Unmarshal([]byte(data[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}
