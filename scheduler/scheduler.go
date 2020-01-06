package scheduler

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
	"github.com/microapis/messages-api"
	"github.com/microapis/messages-api/channel"
	dbBolt "github.com/microapis/messages-api/database/bolt"
	dbRedis "github.com/microapis/messages-api/database/redis"
	pb "github.com/microapis/messages-api/proto"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// StorageConfig is a struct that will be deleted.
type StorageConfig struct {
	RedisHost        string        // URL of the redis server
	RedisPort        string        // URL of the redis server
	RedisLog         bool          // log database commands
	RedisMaxIdle     int           // maximum number of idle connections in the pool
	RedisDatabase    int           // redis database to use
	RedisIdleTimeout time.Duration // timeout for idle connections

	MessageStore *dbBolt.MessageStore
	ChannelStore *dbRedis.ChannelStore
}

// New builds a new messages.Store backed by bolt DB.
//
// In case of any error it panics.
func New(config StorageConfig) messages.SchedulerService {
	s := &service{
		pq:  newPriorityQueue(config),
		idc: make(chan ulid.ULID),

		ms: config.MessageStore,
		cs: config.ChannelStore,
	}

	go s.run()

	return s
}

var msgBucket = []byte("messages")

type service struct {
	db *bolt.DB
	pq *priorityQueue

	idc chan ulid.ULID

	ms *dbBolt.MessageStore
	cs *dbRedis.ChannelStore
}

// Put ...
func (s *service) Put(id ulid.ULID, channel string, provider string, content string, status string) error {
	ch, err := s.cs.Get(channel)
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(ch.Address(), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewMessageBackendServiceClient(conn)
	resp, err := client.Approve(context.Background(), &pb.MessageBackendApproveRequest{Content: content})
	if err != nil {
		// update status to crashed-approve
		e := s.ms.UpdateStatus(id, messages.CrashedApprove)
		if e != nil {
			return e
		}

		// TODO(ca): send callback when could not updated status
		return err
	}
	if !resp.Valid {
		// update status to failed-approve
		err := s.ms.UpdateStatus(id, messages.FailedApprove)
		if err != nil {
			return err
		}

		// TODO(ca): send callback when could not updated status

		if resp.Error != nil {
			return errors.Errorf("invalid message, %s", resp.Error.Message)
		}
		return errors.New("invalid message")
	}

	m := messages.Message{
		ID:       id,
		Content:  content,
		Status:   status,
		Channel:  channel,
		Provider: provider,
	}

	err = s.ms.AddMessage(m)
	if err != nil {
		return err
	}

	s.idc <- id

	return nil
}

// Get ...
func (s *service) Get(id ulid.ULID) (*messages.Message, error) {
	msg, err := s.ms.Get(id)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Update ...
func (s *service) Update(id ulid.ULID, content string) error {
	err := s.ms.UpdateContent(id, content)
	if err != nil {
		return err
	}

	return nil
}

// Cancel ...
func (s *service) Cancel(id ulid.ULID) error {
	ok, err := s.pq.DeleteByID(id)
	if err != nil {
		return err
	}

	if !ok {
		log.Printf("%s not found in priority queue", id)
		return nil
	}

	err = s.ms.UpdateStatus(id, messages.Cancelled)
	if err != nil {
		return err
	}

	return nil
}

// Register ...
func (s *service) Register(c channel.Channel) error {
	err := s.cs.Register(c)
	if err != nil {
		return err
	}

	return nil
}

// Run in its goroutine
func (s *service) run() {
	var next uint64
	var timer *time.Timer

	pq := s.pq
	for {
		var tick <-chan time.Time

		top := pq.Peek()
		if top != nil {
			if t := top.Time(); t < next || next == 0 {
				var delay int64
				now := ulid.Timestamp(time.Now())
				if t >= now {
					delay = int64(t - now)
				}

				if timer == nil {
					timer = time.NewTimer(time.Duration(delay) * time.Millisecond)
				} else {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer = time.NewTimer(time.Duration(delay) * time.Millisecond)
				}
			}
		}

		if timer != nil && top != nil {
			tick = timer.C
		}

		select {
		case <-tick:
			id, err := pq.Pop()
			if err != nil {
				log.Printf(err.Error())
			}

			if id != nil {
				go s.send(*id)
			}
			next = 0
		case id := <-s.idc:
			pq.Push(id)
		}
	}
}

func (s *service) send(id ulid.ULID) {
	msg, err := s.Get(id)
	if err != nil {
		log.Printf("Error: could not get message %s, %v", id, err)
		return
	}

	ch, err := s.cs.Get(msg.Channel)
	if err != nil {
		log.Printf("Error: could not get channel, backend %s is not register, %v", msg.Channel, err)
		return
	}

	// TODO(ja): use secure connections

	conn, err := grpc.Dial(ch.Address(), grpc.WithInsecure())
	if err != nil {
		log.Printf("Error: could not connect to backend at %s, %v", msg.Provider, err)
		return
	}
	defer conn.Close()

	client := pb.NewMessageBackendServiceClient(conn)
	resp, err := client.Deliver(context.Background(), &pb.MessageBackendDeliverRequest{Content: msg.Content})
	if err != nil {
		log.Printf("Error: could not deliver message %s, %v", msg.ID, err)

		// update status to crashed-deliver
		e := s.ms.UpdateStatus(id, messages.CrashedDeliver)
		if e != nil {
			log.Printf("Error: could not update message status %s, %v", msg.ID, err)
			return
		}

		// TODO(ca): send callback when could not updated status
		return
	}
	if resp.Error != nil {
		log.Printf("Error: failed to deliver message %s, %v", msg.ID, resp.Error.Message)

		// update status to failed-deliver
		e := s.ms.UpdateStatus(id, messages.FailedDeliver)
		if e != nil {
			// TODO(ca): check this error
			log.Printf("Error: could not update message status %s, %v", msg.ID, err)
			return
		}

		// TODO(ca): send callback when could not updated status
		return
	}

	e := s.ms.UpdateStatus(id, messages.Sent)
	if e != nil {
		log.Printf("Error: could not update message status %s, %v", msg.ID, err)
		return
	}
}

// TODO(ca): move this to other site.
func generateID(criteriaDelay time.Duration) (*ulid.ULID, error) {
	delay := criteriaDelay

	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	id, err := ulid.New(
		ulid.Timestamp(time.Now().Add(delay)),
		entropy,
	)
	if err != nil {
		//TODO: move this message
		log.Printf("Failed to create message id - %v", err)
		return nil, err
	}

	return &id, nil
}
