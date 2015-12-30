package nsq

import (
	"github.com/gocontrib/pubsub"
	"github.com/nsqio/go-nsq"
)

// Subscription channel.
type sub struct {
	consumers []*nsq.Consumer
	closed    chan bool
	send      chan interface{}
}

func (s *sub) Read() <-chan interface{} {
	return s.send
}

func (s *sub) Close() error {
	go func() {
		s.closed <- true
	}()
	go func() {
		for _, c := range s.consumers {
			c.Stop()
		}
	}()
	return nil
}

func (s *sub) CloseNotify() <-chan bool {
	return s.closed
}

func (s *sub) HandleMessage(msg *nsq.Message) error {
	go func() {
		v, err := pubsub.Unmarshal(msg.Body)
		if err != nil {
			return
		}
		s.send <- v
	}()
	return nil
}
