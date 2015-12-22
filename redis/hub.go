package redis

import (
	"runtime/debug"

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
	var conn, err = openRedisConn()
	if err != nil {
		return nil, err

	}
	return &hub{conn}, nil
}

func openRedisConn() (redis.Conn, error) {
	if len(*redisURL) == 0 {
		return redisurl.Connect()
	}
	return redisurl.ConnectToURL(*redisURL)
}

// PubSub powered by redis
type hub struct {
	conn redis.Conn
}

func (h *hub) Close() error {
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
	var cn, err = openRedisConn()
	if err != nil {
		return nil, err
	}

	var chans []interface{}
	for _, name := range channels {
		chans = append(chans, name)
	}

	s := &sub{
		channels: chans,
		conn:     redis.PubSubConn{Conn: cn},
		closed:   make(chan bool),
		send:     make(chan interface{}),
	}
	go s.start()
	return s, nil
}

// Subscription channel.
type sub struct {
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
		s.conn.Unsubscribe(s.channels...)

		// TODO safe stop of start goroutine
		s.conn.Close()
		close(s.send)
		s.closed <- true
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
			log.Error("%v", m)
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
