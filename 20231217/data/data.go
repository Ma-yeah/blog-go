package data

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

func NewRedis() (redis.UniversalClient, error) {
	c := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        []string{"127.0.0.1:6379"},
		DB:           0,
		MaxRetries:   1,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		PoolSize:     1000,
		MinIdleConns: 100,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return c, nil
}

var GlobalRedis redis.UniversalClient

func init() {
	GlobalRedis, _ = NewRedis()
}
