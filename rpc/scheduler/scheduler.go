package schedulersvc

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/microapis/messages-api"
	"github.com/microapis/messages-api/channel"
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

	if err := s.schedulerSvc.Put(id, channel, provider, content, messages.Pending); err != nil {
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

// Register ...
func (s *Service) Register(ctx context.Context, r *pb.MessageRegisterRequest) (*pb.MessageRegisterResponse, error) {
	c := r.GetChannel()
	n := c.GetName()
	h := c.GetHost()
	p := c.GetPort()

	providers := make([]*channel.Provider, 0)
	for _, v := range c.GetProviders() {
		provider := &channel.Provider{
			Name:   v.GetName(),
			Params: v.GetParams(),
		}

		providers = append(providers, provider)
	}

	channel := channel.Channel{
		Name:      n,
		Providers: providers,
		Host:      h,
		Port:      p,
	}

	err := s.schedulerSvc.Register(channel)
	if err != nil {
		log.Println(fmt.Sprintf("[gRPC][MessagesService][Register][Error] error = %v", err))
		return &pb.MessageRegisterResponse{
			Error: &pb.MessagesError{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	log.Println(fmt.Sprintf("[gRPC][MessagesService][Update][Register]"))
	return &pb.MessageRegisterResponse{}, nil
}
