package nats

import (
	"sync"

	"github.com/gocontrib/pubsub"
	nats "github.com/nats-io/nats.go"
)

// Subscription channel.
type sub struct {
	sync.Mutex
	hub    *hub
	subs   []*nats.Subscription
	send   chan interface{}
	closed chan bool
}

func (s *sub) Read() <-chan interface{} {
	return s.send
}

func (s *sub) Close() error {
	go func() {
		s.Lock()
		defer s.Unlock()

		if s.hub == nil {
			return
		}

		s.hub.remove(s)
		s.hub = nil

		for _, t := range s.subs {
			t.Unsubscribe()
		}

		s.closed <- true
		close(s.send)
	}()
	return nil
}

func (s *sub) CloseNotify() <-chan bool {
	return s.closed
}

func (s *sub) Handler(msg *nats.Msg) {
	go func() {
		v, err := pubsub.Unmarshal(msg.Data)
		if err != nil {
			return
		}
		s.send <- v
	}()
}
