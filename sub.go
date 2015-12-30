package pubsub

// Subscription to multiple hub channels.
type sub struct {
	channels []*channel
	closed   chan bool
	send     chan interface{}
}

func makeSub(channels []*channel) *sub {
	return &sub{
		channels: channels,
		closed:   make(chan bool),
		send:     make(chan interface{}),
	}
}

// Read returns channel of receiver events.
func (s *sub) Read() <-chan interface{} {
	return s.send
}

// Close removes subscriber from channel.
func (s *sub) Close() error {
	go func() {
		for _, c := range s.channels {
			c.unsubscribe <- s
		}
	}()
	go func() {
		s.closed <- true
		close(s.send)
	}()
	return nil
}

// CloseNotify returns channel to handle close event.
func (s *sub) CloseNotify() <-chan bool {
	return s.closed
}
