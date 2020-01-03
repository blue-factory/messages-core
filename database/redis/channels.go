package redis

import (
	"encoding/json"
	"log"

	"github.com/microapis/messages-api"

	db "github.com/microapis/messages-api/database"
)

// ChannelStore ...
type ChannelStore struct {
	Dst *db.RedisDatastore
}

// NewChannelStore ...
func NewChannelStore(dst *db.RedisDatastore) (*ChannelStore, error) {
	return &ChannelStore{
		Dst: dst,
	}, nil
}

// Register ...
func (ss *ChannelStore) Register(c messages.Channel) error {
	// TODO(ca): should get redis c.name value and also merge c.Providers and cc.Providers

	str, err := json.Marshal(c)
	if err != nil {
		return err
	}

	log.Println("str", str)

	err = ss.Dst.Client.Set(c.Name, string(str), 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get ...
func (ss *ChannelStore) Get(name string) (*messages.Channel, error) {
	val, err := ss.Dst.Client.Get(name).Result()
	if err != nil {
		return nil, err
	}

	log.Println("key", val)

	c := &messages.Channel{}
	err = json.Unmarshal([]byte(val), c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetAll ...
func (ss *ChannelStore) GetAll() ([]*messages.Channel, error) {
	keys, err := ss.Dst.Client.Keys("*").Result()
	if err != nil {
		return nil, err
	}

	log.Println("keys", keys)

	values, err := ss.Dst.Client.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}

	log.Println("value", values)

	cc := make([]*messages.Channel, 0)

	for _, v := range values {
		c := &messages.Channel{}
		err = json.Unmarshal([]byte(v.(string)), c)
		if err != nil {
			return nil, err
		}

		cc = append(cc, c)
	}

	return cc, nil
}
