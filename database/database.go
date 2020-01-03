package db

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/go-redis/redis"
)

// BoltDatastore store data in db using bolt as a db backend
type BoltDatastore struct {
	DB *bolt.DB
}

// RedisDatastore store channel data in cache db
type RedisDatastore struct {
	Client *redis.Client
}

var (
	MsgBucket = []byte("messages")
)

// NewBoltDatastore returns a new datastore instance or an error if
// a datasore cannot be returned
func NewBoltDatastore(path string) (*BoltDatastore, error) {
	db, err := bolt.Open(path, os.ModePerm, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, berr := tx.CreateBucketIfNotExists(MsgBucket)
		return berr
	})
	if err != nil {
		return nil, err
	}

	return &BoltDatastore{
		DB: db,
	}, nil
}

// NewRedisDatastore returns a new datastore instance or an error if
// a datasore cannot be returned
func NewRedisDatastore(host, port string) (*RedisDatastore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &RedisDatastore{
		Client: client,
	}, nil
}
