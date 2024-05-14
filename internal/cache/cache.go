package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Cache *redis.Client

func GetCache(ctx context.Context) *redis.Client {
	if Cache == nil || Cache.Ping(ctx).Err() != nil {
		Cache = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	return Cache
}
