package service

import (
	"fmt"
	"log"
	"net"

	"github.com/microapis/messages-core/proto"
	"github.com/microapis/messages-core/scheduler"
	schedulersvc "github.com/microapis/messages-core/scheduler"

	channeldb "github.com/microapis/messages-core/channel/database"
	"github.com/microapis/messages-core/channel/database/redis"

	messagedb "github.com/microapis/messages-core/message/database"
	"github.com/microapis/messages-core/message/database/bolt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServiceConfig ...
type ServiceConfig struct {
	Addr string

	RedisURL string

	Approve func(content string) (bool, error)
	Deliver func(content string) error
}

// Service ...
type Service struct {
	Instance *schedulersvc.Service
	Name     string
	Addr     string
}

// NewMessageService ...
func NewMessageService(name string, config ServiceConfig) (*Service, error) {
	// ----- Init DB
	boltDst, err := messagedb.NewBoltDatastore("messages.db")
	if err != nil {
		return nil, err
	}

	redisDst, err := channeldb.NewRedisDatastore(config.RedisURL)
	if err != nil {
		return nil, err
	}

	// initialize message store
	ms, err := bolt.NewMessageStore(boltDst)
	if err != nil {
		return nil, err
	}

	// initialize channel store
	cs, err := redis.NewChannelStore(redisDst)
	if err != nil {
		return nil, err
	}

	svc := schedulersvc.NewRPC(scheduler.StorageConfig{
		MessageStore: ms,
		ChannelStore: cs,

		RedisURL: config.RedisURL,

		Approve:  config.Approve,
		Delivery: config.Deliver,
	})

	return &Service{
		Instance: svc,
		Name:     name,
		Addr:     config.Addr,
	}, nil
}

// Run ...
func (s *Service) Run() error {
	// initialize gprc server
	srv := grpc.NewServer()

	proto.RegisterSchedulerServiceServer(srv, s.Instance)
	reflection.Register(srv)

	log.Println("Starting Messages " + s.Name + " service...")

	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Messages "+s.Name+" service, Listening on: %v", s.Addr))

	if err := srv.Serve(lis); err != nil {
		return err
	}

	return nil
}
