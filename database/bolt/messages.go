package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/microapis/messages-api"
	"github.com/oklog/ulid"

	db "github.com/microapis/messages-api/database"
	pb "github.com/microapis/messages-api/proto"
)

// MessageStore ...
type MessageStore struct {
	Dst *db.BoltDatastore
}

// NewMessageStore ...
func NewMessageStore(dst *db.BoltDatastore) (*MessageStore, error) {
	return &MessageStore{
		Dst: dst,
	}, nil
}

// AddMessage ...
func (ss *MessageStore) AddMessage(m messages.Message) error {
	err := ss.Dst.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(db.MsgBucket)

		k, merr := m.ID.MarshalBinary()
		if merr != nil {
			return merr
		}
		v, jerr := proto.Marshal(&pb.Message{
			Id:       m.ID.String(),
			Channel:  string(m.Channel),
			Provider: string(m.Provider),
			Content:  string(m.Content),
		})
		if jerr != nil {
			return jerr
		}
		return b.Put(k, v)
	})
	if err != nil {
		return err
	}

	return nil
}

// Get ...
func (ss *MessageStore) Get(id ulid.ULID) (*messages.Message, error) {
	var msg pb.Message
	err := ss.Dst.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(db.MsgBucket)
		k, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		v := b.Get(k)
		if err := proto.Unmarshal(v, &msg); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &messages.Message{
		ID:       id,
		Channel:  msg.Channel,
		Provider: msg.Provider,
		Content:  msg.Content,
	}, nil
}

// UpdateContent ...
func (ss *MessageStore) UpdateContent(id ulid.ULID, content string) error {
	var msg pb.Message
	return ss.Dst.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(db.MsgBucket)
		k, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		v := b.Get(k)
		if err = proto.Unmarshal(v, &msg); err != nil {
			return err
		}
		msg.Content = content
		v, err = proto.Marshal(&msg)
		if err != nil {
			return err
		}
		return b.Put(k, v)
	})
}

// UpdateStatus ...
func (ss *MessageStore) UpdateStatus(id ulid.ULID, status string) error {
	var msg pb.Message
	return ss.Dst.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(db.MsgBucket)
		k, err := id.MarshalBinary()
		if err != nil {
			return err
		}
		v := b.Get(k)
		if err = proto.Unmarshal(v, &msg); err != nil {
			return err
		}
		msg.Status = string(status)
		v, err = proto.Marshal(&msg)
		if err != nil {
			return err
		}
		return b.Put(k, v)
	})
}
