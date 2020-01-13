package db

import (
	"os"

	"github.com/boltdb/bolt"
)

// BoltDatastore store data in db using bolt as a db backend
type BoltDatastore struct {
	DB *bolt.DB
}

var (
	// MsgBucket ...
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
