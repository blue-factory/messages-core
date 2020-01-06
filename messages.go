package messages

import (
	pb "github.com/microapis/messages-api/proto"
	"github.com/microapis/messages-api/channel"
	"github.com/oklog/ulid"
)

// Statuses ...
const (
	// Pending ...
	Pending = "pending"
	// Sent ...
	Sent = "sent"
	// FailedApprove ...
	FailedApprove = "failed-approve"
	// CrashedApprove ...
	CrashedApprove = "crashed-approve"
	// FailedDeliver ...
	FailedDeliver = "failed-deliver"
	// CrashedDeliver ...
	CrashedDeliver = "crashed-deliver"
	// Cancelled ...
	Cancelled = "cancelled"
)

// Message describes a message that needs to be delivered by the system.
type Message struct {
	// ID is an ULID that uniquely identifies (https://github.com/alizain/ulid)
	// a message and encodes the time when the message needs to be sent.
	ID ulid.ULID `json:"id"`

	// Channel identifies the type of message to send
	Channel string `json:"channel"`

	// Content is an arbitrary byte slice that describes the message to
	// be sent.
	//
	// The format of the content varies by the Backend used, and to avoid
	// latter failures the Backend must validate the content before the
	// approval of the message.
	Content string `json:"content,string"`

	// Provider identifies the Backend service used to send the message.
	Provider string `json:"provider"`

	// Status ...
	Status string `json:"status"`
}

// ToProto ...
func (m *Message) ToProto() *pb.Message {
	return &pb.Message{
		Id:       m.ID.String(),
		Content:  m.Content,
		Provider: m.Provider,
		Status:   m.Status,
	}
}

// FromProto ...
func (m *Message) FromProto(mm *pb.Message) (*Message, error) {
	id, err := ulid.Parse(mm.Id)
	if err != nil {
		return nil, err
	}

	m.ID = id
	m.Content = mm.Content
	m.Provider = mm.Provider
	m.Status = mm.Status

	return m, nil
}

// SchedulerService stores and keep track of the statuses of messages.
type SchedulerService interface {
	// Put stores a message content and schedule the delivery on t time.
	// TODO(ca): change subjectID params to ulid.ULID type
	Put(id ulid.ULID, channel string, provider string, content string, status string) error

	// Get retrieves the message with the given id.
	//
	// In case of any error the Message will be nil.
	Get(id ulid.ULID) (*Message, error)

	// Update updates the content of the message with the given id.
	Update(id ulid.ULID, content string) error

	// Cancel cancel the message with the given id.
	Cancel(id ulid.ULID) error

	// Register new channel provider service
	Register(channel channel.Channel) error
}

// Backend manages the approval and delivery of messages.
type Backend interface {
	// Aprove validates the content of a message.
	//
	// If the message is valid the error will be nil, otherwise the error
	// must be non-nil and describe why the message is invalid.
	Approve(content string) (ok bool, err error)

	// Deliver delivers the message encoded in content.
	Deliver(content string) error
}

// Email ...
type Email struct {
	From     string   `json:"from"`
	FromName string   `json:"from_name"`
	To       []string `json:"to"`
	ReplyTo  []string `json:"reply_to"`
	Subject  string   `json:"subject"`
	Text     string   `json:"text"`
	HTML     string   `json:"html"`
	Provider string   `json:"provider"`
}

// SMS ...
type SMS struct {
	Phone    string `json:"phone"`
	Text     string `json:"text"`
	Provider string `json:"provider"`
}
