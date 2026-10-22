package configs

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	rdbHost := os.Getenv("REDIS_HOST")
	if rdbHost == "" {
		rdbHost = os.Getenv("RDBHOST")
	}
	if rdbHost == "" {
		rdbHost = "127.0.0.1"
	}

	rdbPort := os.Getenv("REDIS_PORT")
	if rdbPort == "" {
		rdbPort = os.Getenv("RDBPORT")
	}
	if rdbPort == "" {
		rdbPort = "6379"
	}

	rdbPass := os.Getenv("REDIS_PASSWORD")
	if rdbPass == "" {
		rdbPass = os.Getenv("RDBPASS")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rdbHost, rdbPort),
		Password: rdbPass,
		DB:       0,
	})

	return client
}

func PingRedis(rdb *redis.Client) error {
	return rdb.Ping(context.Background()).Err()
}
