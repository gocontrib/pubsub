package redis

import (
	"os"
	"runtime/debug"
	"sync"

	"github.com/drone/config"
	"github.com/garyburd/redigo/redis"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/soveran/redisurl"
)

var (
	redisURL = config.String("pubsub-redis", "")
)

func init() {
	pubsub.RegisterDriver(&driver{}, "redis")
}

type driver struct{}

func (d *driver) Create() (pubsub.Hub, error) {
	log.Info("connecting to redis pubsub")
	return Open()
}

// Open creates pubsub hub connected to redis server.
func Open(URL ...string) (pubsub.Hub, error) {
	redisURL := getRedisURL(URL...)
	conn, err := redisurl.ConnectToURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &hub{
		conn:     conn,
		redisURL: redisURL,
		subs:     make(map[*sub]struct{}),
	}, nil
}

func getRedisURL(URL ...string) string {
	if len(URL) == 1 && len(URL[0]) > 0 {
		return URL[0]
	}
	if len(*redisURL) == 0 {
		return os.Getenv("REDIS_URL")
	}
	return *redisURL
}

// PubSub powered by redis
type hub struct {
	sync.Mutex
	conn     redis.Conn
	redisURL string
	subs     map[*sub]struct{}
}

func (h *hub) Close() error {
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

		s.conn.Unsubscribe(s.channels...)

		// TODO safe stop of start goroutine
		s.conn.Close()

		s.closed <- true
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
			log.Errorf("redis pubsub error: %+v", m)
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
