package message

import (
	"github.com/microapis/messages-core/proto"
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
func (m *Message) ToProto() *proto.Message {
	return &proto.Message{
		Id:       m.ID.String(),
		Content:  m.Content,
		Provider: m.Provider,
		Status:   m.Status,
	}
}

// FromProto ...
func (m *Message) FromProto(mm *proto.Message) (*Message, error) {
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
