package pubsub

import "time"

// Hub interface of pubsub system.
type Hub interface {
	// Publish sends input message to specified channels.
	Publish(channels []string, msg interface{})
	// Subscribe opens channel to listen specified channels.
	Subscribe(channels []string) (Channel, error)
	// Close stops the pubsub hub.
	Close() error
}

// Channel to listen pubsub events.
type Channel interface {
	// Read returns channel to receive events.
	Read() <-chan interface{}
	// Close stops listening underlying pubsub channels.
	Close() error
	// CloseNotify returns channel to receive event when this channel is closed.
	CloseNotify() <-chan bool
}

// Event defines abstract event usually about HTTP update
type Event struct {
	ID           string      `json:"id,omitempty"`     // event id
	Action       string      `json:"action,omitempty"` // http method or specific action
	Method       string      `json:"method,omitempty"` // http method
	URL          string      `json:"url,omitempty"`
	ResourceID   string      `json:"resource_id,omitempty"`   // resource id
	ResourceType string      `json:"resource_type,omitempty"` // resource type
	Payload      interface{} `json:"payload,omitempty"`       // input payload
	CreatedBy    string      `json:"created_by,omitempty"`
	CreatedAt    time.Time   `json:"created_at,omitempty"`
	Result       interface{} `json:"result,omitempty"` // any result of mutation
}
