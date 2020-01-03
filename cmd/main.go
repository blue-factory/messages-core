package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/microapis/messages-api/database/bolt"
	"github.com/microapis/messages-api/database/redis"

	schedulersvc "github.com/microapis/messages-api/rpc/scheduler"

	db "github.com/microapis/messages-api/database"
	"github.com/microapis/messages-api/proto"
	"github.com/microapis/messages-api/scheduler"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	dbfile := flag.String("db", "messages.db", "file to store messages")

	redisIdleTimeout := flag.Duration("redis_idle_timeout", 5*time.Second, "Timeout for redis idle connections.")
	redisDatabase := flag.Int("redis_db", 1, "Redis database to use")
	redisMaxIdle := flag.Int("redis_max_idle", 10, "Maximum number of idle connections in the pool")

	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		err := errors.New("invalid PORT env value")
		log.Fatal(err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		err := errors.New("invalid REDIS_HOST env value")
		log.Fatal(err)
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		err := errors.New("invalid REDIS_PORT env value")
		log.Fatal(err)
	}

	// ----- Init DB
	boltDst, err := db.NewBoltDatastore(*dbfile)
	if err != nil {
		log.Fatal(err)
	}
	redisDst, err := db.NewRedisDatastore(redisHost, redisPort)
	if err != nil {
		log.Fatal(err)
	}

	// initialize message store
	ms, err := bolt.NewMessageStore(boltDst)
	if err != nil {
		log.Fatal(err)
	}

	// initialize channel store
	cs, err := redis.NewChannelStore(redisDst)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("%s:%s", "0.0.0.0", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	// ----- Init grpc
	s := grpc.NewServer()
	log.Printf("Starting server at %s redis_url: %s redis_db: %d database: %s \n", redisHost, redisPort, *redisDatabase, *dbfile)
	proto.RegisterMessageServiceServer(s, schedulersvc.New(scheduler.StorageConfig{
		MessageStore:     ms,
		ChannelStore:     cs,
		RedisHost:        redisHost,
		RedisPort:        redisPort,
		RedisIdleTimeout: *redisIdleTimeout,
		RedisDatabase:    *redisDatabase,
		RedisMaxIdle:     *redisMaxIdle,
	}))

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
