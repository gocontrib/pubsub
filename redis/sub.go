package redis

import (
	"runtime/debug"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
)

// Subscription channel.
type sub struct {
	sync.Mutex
	hub      *hub
	channels []interface{}
	conn     redis.PubSubConn
	closed   chan bool
	send     chan interface{}
}

// Read returns channel of receiver events.
func (s *sub) Read() <-chan interface{} {
	return s.send
}

// Close removes subscriber from channel.
func (s *sub) Close() error {
	go func() {
		s.Lock()
		defer s.Unlock()

		if s.hub == nil {
			return
		}

		s.hub.remove(s)
		s.hub = nil
		s.closed <- true

		s.conn.Unsubscribe(s.channels...)

		// TODO safe stop of start goroutine
		s.conn.Close()

		close(s.send)
	}()
	return nil
}

// CloseNotify returns channel to handle close event.
func (s *sub) CloseNotify() <-chan bool {
	return s.closed
}

func (s *sub) start() {
	if err := recover(); err != nil {
		log.Errorf("recovered from panic: %+v", err)
		debug.PrintStack()
	}

	s.conn.Subscribe(s.channels...)

	for {
		switch m := s.conn.Receive().(type) {
		case redis.Message:
			s.push(m.Data)
		case redis.PMessage:
			s.push(m.Data)
		case redis.Subscription:
			log.Info("Subscription: %s %s %d", m.Kind, m.Channel, m.Count)
			if m.Count == 0 {
				return
			}
		case error:
			return
		}
	}
}

func (s *sub) push(data []byte) {
	go func() {
		v, err := pubsub.Unmarshal(data)
		if err != nil {
			return
		}
		s.send <- v
	}()
}
