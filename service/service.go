package service

import (
	"fmt"
	"log"
	"net"

	"github.com/microapis/messages-lib/proto"
	"github.com/microapis/messages-lib/scheduler"
	schedulersvc "github.com/microapis/messages-lib/scheduler"

	channeldb "github.com/microapis/messages-lib/channel/database"
	"github.com/microapis/messages-lib/channel/database/redis"

	messagedb "github.com/microapis/messages-lib/message/database"
	"github.com/microapis/messages-lib/message/database/bolt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServiceConfig ...
type ServiceConfig struct {
	Port string

	RedisHost     string
	RedisPort     string
	RedisDatabase int

	Approve func(content string) (bool, error)
	Deliver func(content string) error
}

// Service ...
type Service struct {
	Instance *schedulersvc.Service
	Name     string
	Port     string
}

// NewMessageService ...
func NewMessageService(name string, config ServiceConfig) (*Service, error) {
	// ----- Init DB
	boltDst, err := messagedb.NewBoltDatastore("messages.db")
	if err != nil {
		return nil, err
	}

	redisDst, err := channeldb.NewRedisDatastore(config.RedisHost, config.RedisPort)
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

		RedisHost:     config.RedisHost,
		RedisPort:     config.RedisPort,
		RedisDatabase: config.RedisDatabase,

		Approve:  config.Approve,
		Delivery: config.Deliver,
	})

	return &Service{
		Instance: svc,
		Name:     name,
		Port:     config.Port,
	}, nil
}

// Run ...
func (s *Service) Run() error {
	// define address value to grpc service
	addr := fmt.Sprintf("0.0.0.0:%s", s.Port)

	// initialize gprc server
	srv := grpc.NewServer()

	proto.RegisterSchedulerServiceServer(srv, s.Instance)
	reflection.Register(srv)

	log.Println("Starting Messages " + s.Name + " service...")

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Messages "+s.Name+" service, Listening on: %v", s.Port))

	if err := srv.Serve(lis); err != nil {
		return err
	}

	return nil
}
