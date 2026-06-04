package rdb

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func RedisInit() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Println("empty", addr)
	}
	Client = redis.NewClient(&redis.Options{Addr: addr})
	return Client
}

func RedisClose() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

func RedisPing(ctx context.Context) error {
	return Client.Ping(ctx).Err()
}
