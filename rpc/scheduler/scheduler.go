package schedulersvc

import (
	"log"
	"math/rand"
	"time"

	"github.com/microapis/messages-api"
	"github.com/microapis/messages-api/scheduler"
	"golang.org/x/net/context"

	pb "github.com/microapis/messages-api/proto"
	"github.com/oklog/ulid"
)

var _ pb.MessageServiceServer = (*Service)(nil)

// Service ...
type Service struct {
	schedulerSvc messages.SchedulerService
}

// New ...
func New(config scheduler.StorageConfig) *Service {
	return &Service{
		schedulerSvc: scheduler.New(config),
	}
}

// Put ...
func (s *Service) Put(ctx context.Context, r *pb.MessagePutRequest) (*pb.MessagePutResponse, error) {
	delay := r.GetDelay()
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))

	id, err := ulid.New(
		ulid.Timestamp(time.Now().Add(time.Duration(delay)*time.Second)),
		entropy,
	)
	if err != nil {
		log.Println("Failed to create message id", err)
		return &pb.MessagePutResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	channel := r.GetChannel()
	provider := r.GetProvider()
	content := r.GetContent()

	if err := s.schedulerSvc.Put(id, channel, provider, content, messages.Pending); err != nil {
		return &pb.MessagePutResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.MessagePutResponse{
		Data: &pb.MessagePutDataResponse{
			Id: id.String(),
		},
	}, nil
}

// Get ...
func (s *Service) Get(ctx context.Context, r *pb.MessageGetRequest) (*pb.MessageGetResponse, error) {
	id, err := ulid.Parse(r.GetId())
	if err != nil {
		return &pb.MessageGetResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	msg, err := s.schedulerSvc.Get(id)
	if err != nil {
		return &pb.MessageGetResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.MessageGetResponse{
		Data: &pb.Message{
			Id:       r.Id,
			Content:  string(msg.Content),
			Provider: string(msg.Provider),
			Status:   string(msg.Status),
		},
	}, nil
}

// Update ...
func (s *Service) Update(ctx context.Context, r *pb.MessageUpdateRequest) (*pb.MessageUpdateResponse, error) {
	id, err := ulid.Parse(r.GetId())
	if err != nil {
		return &pb.MessageUpdateResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	content := r.GetContent()

	if err := s.schedulerSvc.Update(id, content); err != nil {
		return &pb.MessageUpdateResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.MessageUpdateResponse{}, nil
}

// Cancel ...
func (s *Service) Cancel(ctx context.Context, r *pb.MessageCancelRequest) (*pb.MessageCancelResponse, error) {
	id, err := ulid.Parse(r.GetId())
	if err != nil {
		return &pb.MessageCancelResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	if err := s.schedulerSvc.Cancel(id); err != nil {
		return &pb.MessageCancelResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.MessageCancelResponse{}, nil
}

// Register ...
func (s *Service) Register(ctx context.Context, r *pb.MessageRegisterRequest) (*pb.MessageRegisterResponse, error) {
	c := r.GetChannel()
	providers := make([]*messages.Provider, 0)

	for _, v := range c.GetProviders() {
		provider := &messages.Provider{
			Name:   v.GetName(),
			Params: v.GetParams(),
		}

		providers = append(providers, provider)
	}

	channel := messages.Channel{
		Name:      c.GetName(),
		Providers: providers,
		Host:      c.GetHost(),
		Port:      c.GetPort(),
	}

	err := s.schedulerSvc.Register(channel)
	if err != nil {
		return &pb.MessageRegisterResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.MessageRegisterResponse{}, nil
}
