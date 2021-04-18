package tinyurl

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func Cache(short, origin string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
	})
	pipe := rdb.Pipeline()
	ctx := context.TODO()
	pipe.Auth(ctx, viper.GetString("redis.password"))

	return pipe.Set(ctx, "tiny:"+short, origin, 24*time.Hour).Err()
}

func CacheGet(short string) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
	})
	pipe := rdb.Pipeline()
	ctx := context.TODO()
	pipe.Auth(ctx, viper.GetString("redis.password"))

	return pipe.Get(ctx, "tiny:"+short).Result()
}
