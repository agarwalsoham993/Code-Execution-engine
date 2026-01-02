package queue

import (
	"code-runner/pkg/models"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
	key    string
}

func NewRedisQueue(addr, pwd string) *RedisQueue {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       0,
	})
	return &RedisQueue{
		client: rdb,
		ctx:    context.Background(),
		key:    "execution_queue",
	}
}

func (q *RedisQueue) Enqueue(payload models.JobPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return q.client.RPush(q.ctx, q.key, data).Err()
}

func (q *RedisQueue) Dequeue(timeout time.Duration) (*models.JobPayload, error) {
	res, err := q.client.BLPop(q.ctx, timeout, q.key).Result()
	if err != nil {
		return nil, err
	}
	// res[0] is key, res[1] is value
	payload := new(models.JobPayload)
	if err := json.Unmarshal([]byte(res[1]), payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (q *RedisQueue) Length() (int64, error) {
	return q.client.LLen(q.ctx, q.key).Result()
}