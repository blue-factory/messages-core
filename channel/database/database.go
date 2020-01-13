package db

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

// RedisDatastore store channel data in cache db
type RedisDatastore struct {
	Client *redis.Client
}

// NewRedisDatastore returns a new datastore instance or an error if
// a datasore cannot be returned
func NewRedisDatastore(host, port string) (*RedisDatastore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		DB:   0, // channels db
	})

	log.Println("Redis: ping")
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	log.Println("Redis: pong")

	return &RedisDatastore{
		Client: client,
	}, nil
}
