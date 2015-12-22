package pubsub

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
