package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisDB() (*redis.Client, error) {
	options := &redis.Options{
		Addr:     "localhost:32771",
		Password: "",
		DB:       0,
	}
	redisDB := redis.NewClient(options)
	err := redisDB.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return redisDB, nil
}
