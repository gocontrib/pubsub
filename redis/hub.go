package redis

import (
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/gocontrib/pubsub"
	"github.com/soveran/redisurl"
)

// pubsub hub powered by redis
type hub struct {
	sync.Mutex
	conn     redis.Conn
	redisURL string
	subs     map[*sub]struct{}
}

func (h *hub) Close() error {
	h.Lock()
	defer h.Unlock()
	for s := range h.subs {
		s.Close()
	}
	return h.conn.Close()
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
		for _, name := range channels {
			h.conn.Do("PUBLISH", name, data)
		}
	}()
}

func (h *hub) Subscribe(channels []string) (pubsub.Channel, error) {
	cn, err := redisurl.ConnectToURL(h.redisURL)
	if err != nil {
		return nil, err
	}

	var chans []interface{}
	for _, name := range channels {
		chans = append(chans, name)
	}

	h.Lock()
	defer h.Unlock()

	s := &sub{
		hub:      h,
		channels: chans,
		conn:     redis.PubSubConn{Conn: cn},
		closed:   make(chan bool),
		send:     make(chan interface{}),
	}
	h.subs[s] = struct{}{}

	go s.start()
	return s, nil
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
