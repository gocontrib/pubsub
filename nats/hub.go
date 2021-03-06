package nats

import (
	"sync"

	"github.com/gocontrib/pubsub"
	nats "github.com/nats-io/nats.go"
)

// pubsub.Hub impl

type hub struct {
	sync.Mutex
	conn *nats.Conn
	subs map[*sub]struct{}
}

func (h *hub) Publish(channels []string, msg interface{}) {
	if len(channels) == 0 {
		return
	}
	go func() {
		var data, err = pubsub.Marshal(msg)
		if err != nil {
			return
		}
		for _, cn := range channels {
			h.conn.Publish(cn, data)
		}
	}()
}

func (h *hub) Subscribe(channels []string) (pubsub.Channel, error) {

	s := &sub{
		hub:    h,
		send:   make(chan interface{}),
		closed: make(chan bool),
	}

	for _, subject := range channels {
		sub, err := h.conn.Subscribe(subject, s.Handler)
		if err != nil {
			s.Close()
			return nil, err
		}
		s.subs = append(s.subs, sub)
	}

	h.Lock()
	defer h.Unlock()
	h.subs[s] = struct{}{}

	return s, nil
}

func (h *hub) Close() error {
	h.Lock()
	defer h.Unlock()
	for s := range h.subs {
		s.Close()
	}
	h.conn.Close()
	return nil
}

func (h *hub) remove(s *sub) bool {
	h.Lock()
	defer h.Unlock()
	_, ok := h.subs[s]
	if !ok {
		return false
	}
	delete(h.subs, s)
	return true
}
