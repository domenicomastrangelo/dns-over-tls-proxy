package cache

import (
	"fmt"

	"dns-over-tls-proxy/internal/config"

	"github.com/redis/go-redis/v9"
)

var Cache *redis.Client

func GetCache(config config.Config) *redis.Client {
	if Cache == nil || Cache.Ping(config.Ctx).Err() != nil {
		Cache = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
			Password: "",
			DB:       0,
		})
	}

	return Cache
}
