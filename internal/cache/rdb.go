package cache

import (
	"OilStore/internal/logger"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var Client *redis.Client

func RedisInit(addr string) *redis.Client {
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

func DeleteAllCache(ctx context.Context, key string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := Client.Scan(ctx, cursor, key, 100).Result()
		if err != nil {
			logger.Log.Error("check cache error", zap.Error(err))
		}
		if len(keys) > 0 {
			if err := Client.Del(ctx, keys...).Err(); err != nil {
				logger.Log.Error("delete cahce error", zap.Error(err))
			}
			logger.Log.Info("cache was success deleted", zap.Strings("key", keys))
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
