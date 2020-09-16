package db

import (
	"github.com/go-redis/redis"
)

// RedisDatastore store channel data in cache db
type RedisDatastore struct {
	Client *redis.Client
}

// NewRedisDatastore returns a new datastore instance or an error if
// a datasore cannot be returned
func NewRedisDatastore(url string) (*RedisDatastore, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisDatastore{
		Client: client,
	}, nil
}
