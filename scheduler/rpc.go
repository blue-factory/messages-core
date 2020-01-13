package scheduler

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/microapis/messages-lib/message"
	"golang.org/x/net/context"

	pb "github.com/microapis/messages-lib/proto"
	"github.com/oklog/ulid"
)

var _ pb.SchedulerServiceServer = (*Service)(nil)

// Service ...
type Service struct {
	schedulerSvc SchedulerService
}

// NewRPC ...
func NewRPC(config StorageConfig) *Service {
	return &Service{
		schedulerSvc: New(config),
	}
}

// Put ...
func (s *Service) Put(ctx context.Context, r *pb.MessagePutRequest) (*pb.MessagePutResponse, error) {
	channel := r.GetChannel()
	provider := r.GetProvider()
	content := r.GetContent()
	delay := r.GetDelay()

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Put][Request] channel = %v provider = %v delay = %v", channel, provider, delay))

	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	id, err := ulid.New(
		ulid.Timestamp(time.Now().Add(time.Duration(delay)*time.Second)),
		entropy,
	)
	if err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Put][Error] error = %v", err))
		return &pb.MessagePutResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	if err := s.schedulerSvc.Put(id, channel, provider, content, message.Pending); err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Put][Error] error = %v", err))
		return &pb.MessagePutResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Put][Response] id = %v", id.String()))
	return &pb.MessagePutResponse{
		Data: &pb.MessagePutDataResponse{
			Id: id.String(),
		},
	}, nil
}

// Get ...
func (s *Service) Get(ctx context.Context, r *pb.MessageGetRequest) (*pb.MessageGetResponse, error) {
	log.Println(fmt.Sprintf("[gRPC][MessagesService][Get][Request] id = %v", r.GetId()))

	id, err := ulid.Parse(r.GetId())
	if err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Get][Error] error = %v", err))
		return &pb.MessageGetResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	msg, err := s.schedulerSvc.Get(id)
	if err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Get][Error] error = %v", err))
		return &pb.MessageGetResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Get][Response] id = %v", id.String()))
	return &pb.MessageGetResponse{
		Data: &pb.Message{
			Id:       r.Id,
			Content:  string(msg.Content),
			Channel:  string(msg.Channel),
			Provider: string(msg.Provider),
			Status:   string(msg.Status),
		},
	}, nil
}

// Update ...
func (s *Service) Update(ctx context.Context, r *pb.MessageUpdateRequest) (*pb.MessageUpdateResponse, error) {
	id := r.GetId()
	content := r.GetContent()

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Update][Request] id = %v content = %v", id, content))

	uid, err := ulid.Parse(r.GetId())
	if err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Update][Error] error = %v", err))
		return &pb.MessageUpdateResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	if err := s.schedulerSvc.Update(uid, content); err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Update][Error] error = %v", err))
		return &pb.MessageUpdateResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Update][Response]"))
	return &pb.MessageUpdateResponse{}, nil
}

// Cancel ...
func (s *Service) Cancel(ctx context.Context, r *pb.MessageCancelRequest) (*pb.MessageCancelResponse, error) {
	log.Println(fmt.Sprintf("[gRPC][MessagesService][Update][Request] id = %v", r.GetId()))

	id, err := ulid.Parse(r.GetId())
	if err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Cancel][Error] error = %v", err))
		return &pb.MessageCancelResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	if err := s.schedulerSvc.Cancel(id); err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Cancel][Error] error = %v", err))
		return &pb.MessageCancelResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Cancel][Response]"))
	return &pb.MessageCancelResponse{}, nil
}
