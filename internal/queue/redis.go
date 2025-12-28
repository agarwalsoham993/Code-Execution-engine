package queue

import (
	"context"
	"github.com/redis/go-redis/v9"
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

func (q *RedisQueue) Enqueue(submissionID string) error {
	return q.client.RPush(q.ctx, q.key, submissionID).Err()
}

func (q *RedisQueue) Dequeue() (string, error) {
	// Blocking pop with 0 timeout (wait forever)
	res, err := q.client.BLPop(q.ctx, 0, q.key).Result()
	if err != nil {
		return "", err
	}
	// BLPop returns [key, value]
	return res[1], nil
}
