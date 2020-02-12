package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Network:            "tcp",
		Addr:               "127.0.0.1:6379",
		Dialer:             nil,
		OnConnect:          nil,
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        time.Millisecond * 200,
		ReadTimeout:        time.Millisecond * 5000,
		WriteTimeout:       time.Millisecond * 5000,
		PoolSize:           0,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
	})

	val1 := client.Incr("count_1").Val()
	val2 := client.Incr("count_1").Val()
	val3 := client.IncrBy("count_1", 2).Val()
	count := client.Get("count_1").Val()
	fmt.Println(val1, val2, val3, count)
}
