package cache

import (
	"github.com/go-redis/redis"
)

func NewRedisClient(host, port string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})
	return client
}
