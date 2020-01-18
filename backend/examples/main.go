package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/microapis/messages-core/backend"
)

type service struct{}

func (s *service) Approve(content []byte) (valid bool, err error) {
	if content == nil {
		return false, errors.New("Invalid message content")
	}
	return true, nil
}

func (s *service) Deliver(content []byte) error {
	log.Printf("message received: %s", content)
	return nil
}

func main() {
	host := flag.String("host", "localhost", "host of the service")
	port := flag.Int("port", 5000, "host of the service")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)

	log.Printf("Serving at %s", addr)
	if err := backend.ListenAndServe(addr, &service{}); err != nil {
		log.Fatal(err)
	}
}
