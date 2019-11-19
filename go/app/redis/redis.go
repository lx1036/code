package main

import "github.com/go-redis/redis/v7"

func main() {
	client := redis.NewClient(&redis.Options{
		Network:            "tcp",
		Addr:               "127.0.0.1:7000",
		Dialer:             nil,
		OnConnect:          nil,
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        5,
		ReadTimeout:        1,
		WriteTimeout:       0,
		PoolSize:           0,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
	})
	
	client.Get("test").Val()
}
