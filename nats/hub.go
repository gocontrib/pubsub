package nats

import (
	"sync"

	"github.com/drone/config"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nats-io/nats"
)

var (
	NATS_URL = config.String("pubsub-nats", "")
)

func init() {
	pubsub.RegisterDriver(&driver{}, "nats", "natsio")
}

type driver struct{}

func (d *driver) Create() (pubsub.Hub, error) {
	log.Info("connecting to nats hub")

	var url = *NATS_URL
	if len(url) == 0 {
		url = nats.DefaultURL
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &hub{conn: conn}, nil
}

// pubsub.Hub impl

type hub struct {
	sync.Mutex
	conn *nats.Conn
	subs []*receiver
}

func (h *hub) Publish(channels []string, msg interface{}) {
	if len(channels) == 0 {
		return
	}
	go func() {
		var data, err = pubsub.Marshal(msg)
		if err != nil {
			log.Error("json.Marshal failed: %+v", err)
			return
		}
		for _, cn := range channels {
			h.conn.Publish(cn, data)
		}
	}()
}

func (h *hub) Subscribe(channels []string) (pubsub.Receiver, error) {

	var r = &receiver{
		hub:         h,
		send:        make(chan interface{}),
		closeNotify: make(chan bool),
	}

	for _, subject := range channels {
		sub, err := h.conn.Subscribe(subject, r.Handler)
		if err != nil {
			r.Close()
			return nil, err
		}
		r.subs = append(r.subs, sub)
	}

	h.Lock()
	defer h.Unlock()
	h.subs = append(h.subs, r)

	return r, nil
}

func (h *hub) Close() error {
	for _, r := range h.subs {
		r.Close()
	}
	h.conn.Close()
	return nil
}

func (h *hub) remove(r *receiver) bool {
	for i, t := range h.subs {
		if t == r {
			h.Lock()
			defer h.Unlock()
			h.subs = append(h.subs[:i], h.subs[i+1:]...)
			return true
		}
	}
	return false
}

// pubsub.Receiver impl

type receiver struct {
	sync.Mutex
	hub         *hub
	subs        []*nats.Subscription
	send        chan interface{}
	closeNotify chan bool
	closed      bool
}

func (r *receiver) Read() <-chan interface{} {
	return r.send
}

func (r *receiver) Close() error {
	go func() {
		r.Lock()
		defer r.Unlock()

		if r.closed {
			return
		}

		r.closed = true
		r.hub.remove(r)

		for _, s := range r.subs {
			s.Unsubscribe()
		}

		close(r.send)
		r.closeNotify <- true
	}()
	return nil
}

func (r *receiver) CloseNotify() <-chan bool {
	return r.closeNotify
}

func (r *receiver) Handler(msg *nats.Msg) {
	go func() {
		v, err := pubsub.Unmarshal(msg.Data)
		if err != nil {
			return
		}
		r.send <- v
	}()
}
