package nats

import (
	"sync"

	"github.com/drone/config"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nats-io/nats"
)

var (
	natsURL = config.String("pubsub-nats", "")
)

func init() {
	pubsub.RegisterDriver(&driver{}, "nats", "natsio")
}

type driver struct{}

func (d *driver) Create() (pubsub.Hub, error) {
	log.Info("connecting to nats hub")
	return Open()
}

// Open creates pubsub hub connected to nats server.
func Open() (pubsub.Hub, error) {
	var url = *natsURL
	if len(url) == 0 {
		url = nats.DefaultURL
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &hub{
		conn: conn,
		subs: make(map[*sub]struct{}),
	}, nil
}

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
	for s := range h.subs {
		s.Close()
	}
	h.conn.Close()
	return nil
}

func (h *hub) remove(s *sub) bool {
	_, ok := h.subs[s]
	if !ok {
		return false
	}
	h.Lock()
	defer h.Unlock()
	delete(h.subs, s)
	return true
}

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
