package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"time"
)

type RedisClient struct {
	Rdb *redis.Client
}

func New(dbNum int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       dbNum,
	})

	if rdb == nil {
		return nil, fmt.Errorf("couldn't connect to redis")
	}

	return &RedisClient{Rdb: rdb}, nil
}

func (r *RedisClient) CheckHashedUrl(name string) (string, error) {
	const op = "cache.CheckHashedUrl"

	hashedUrl, err := r.Rdb.Get(name).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}

		return "", fmt.Errorf("%s:%w", op, err)
	}

	return hashedUrl, nil
}

func (r *RedisClient) SaveHashedUrl(name, hashedUrl string) error {
	const op = "cache.SaveTextUrl"

	if err := r.Rdb.Set(name, hashedUrl, 30*time.Second).Err(); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
