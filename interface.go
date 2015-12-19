package pubsub

// Hub interface of pubsub system.
type Hub interface {
	Publish(channels []string, msg interface{})
	Subscribe(channels []string) (Receiver, error)
	Close() error
}

// Receiver of pubsub events.
type Receiver interface {
	Read() <-chan interface{}
	Close() error
	CloseNotify() <-chan bool
}
