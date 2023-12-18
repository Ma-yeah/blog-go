package data

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"time"
)

func newRedis() redis.UniversalClient {
	c := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        viper.GetStringSlice("data.redis.addrs"),
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
		panic(err)
	}
	return c
}

type Repo struct {
	Redis redis.UniversalClient
}

func NewRepo() *Repo {
	return &Repo{
		Redis: newRedis(),
	}
}
