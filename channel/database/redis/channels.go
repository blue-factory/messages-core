package redis

import (
	"context"
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

	ctx := context.Background()
	err = ss.Dst.Client.Set(ctx, c.Name, string(b), 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get ...
func (ss *ChannelStore) Get(name string) (*channel.Channel, error) {
	ctx := context.Background()
	val, err := ss.Dst.Client.Get(ctx, name).Result()
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
	ctx := context.Background()
	keys, err := ss.Dst.Client.Keys(ctx, "*").Result()
	if err == nil {
		return nil, err
	}

	log.Println("keys", keys)

	values, err := ss.Dst.Client.MGet(ctx, keys...).Result()
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
