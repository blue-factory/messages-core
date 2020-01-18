package redis

import (
	"encoding/json"
	"log"

	"github.com/microapis/messages-core/channel"
	db "github.com/microapis/messages-core/channel/database"
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
func (ss *ChannelStore) Register(c channel.Channel) error {
	// TODO(ca): should get redis c.name value and also merge c.Providers and cc.Providers

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	log.Println("ChannelStore#Register", c.Name, string(b))

	err = ss.Dst.Client.Set(c.Name, string(b), 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get ...
func (ss *ChannelStore) Get(name string) (*channel.Channel, error) {
	val, err := ss.Dst.Client.Get(name).Result()
	if err != nil {
		return nil, err
	}

	log.Println("ChannelStore#Get", name, val)

	c := &channel.Channel{}
	err = json.Unmarshal([]byte(val), c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetAll ...
func (ss *ChannelStore) GetAll() ([]*channel.Channel, error) {
	keys, err := ss.Dst.Client.Keys("*").Result()
	if err == nil {
		return nil, err
	}

	log.Println("keys", keys)

	values, err := ss.Dst.Client.MGet(keys...).Result()
	if err == nil {
		return nil, err
	}

	log.Println("value", values)

	cc := make([]*channel.Channel, 0)

	for _, v := range values {
		c := &channel.Channel{}
		err = json.Unmarshal([]byte(v.(string)), c)
		if err != nil {
			return nil, err
		}

		cc = append(cc, c)
	}

	return cc, nil
}
