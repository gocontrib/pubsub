package pubsub

// Hub interface of pubsub system.
type Hub interface {
	Publish(channels []string, msg interface{})
	Subscribe(channels []string) (Channel, error)
	Close() error
}

// Channel to listen pubsub events.
type Channel interface {
	Read() <-chan interface{}
	Close() error
	CloseNotify() <-chan bool
}
