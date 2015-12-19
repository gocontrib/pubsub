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
	REDIS_URL = config.String("pubsub-redis", "")
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
	if len(*REDIS_URL) == 0 {
		return redisurl.Connect()
	}
	return redisurl.ConnectToURL(*REDIS_URL)
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
			log.Error("cannot publish as JSON %v", err)
			return
		}
		for _, name := range channels {
			h.conn.Do("PUBLISH", name, data)
		}
	}()
}

func (h *hub) Subscribe(channels []string) (pubsub.Receiver, error) {
	var cn, err = openRedisConn()
	if err != nil {
		return nil, err
	}

	var chans []interface{}
	for _, name := range channels {
		chans = append(chans, name)
	}

	var r = &receiver{
		channels: chans,
		conn:     redis.PubSubConn{cn},
		closed:   make(chan bool),
		send:     make(chan interface{}),
	}
	go r.start()
	return r, nil
}

// Subscriber channel.
type receiver struct {
	channels []interface{}
	conn     redis.PubSubConn
	closed   chan bool
	send     chan interface{}
}

// Read returns channel of receiver events.
func (r *receiver) Read() <-chan interface{} {
	return r.send
}

// Close removes subscriber from channel.
func (r *receiver) Close() error {
	go func() {
		r.conn.Unsubscribe(r.channels...)

		// TODO safe stop of start goroutine
		r.conn.Close()
		close(r.send)
		r.closed <- true
	}()
	return nil
}

// CloseNotify returns channel to handle close event.
func (r *receiver) CloseNotify() <-chan bool {
	return r.closed
}

func (r *receiver) start() {
	if err := recover(); err != nil {
		log.Error("recovered from panic: %+v", err)
		debug.PrintStack()
	}

	r.conn.Subscribe(r.channels...)

	for {
		switch m := r.conn.Receive().(type) {
		case redis.Message:
			r.push(m.Data)
		case redis.PMessage:
			r.push(m.Data)
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

func (r *receiver) push(data []byte) {
	go func() {
		v, err := pubsub.Unmarshal(data)
		if err != nil {
			return
		}
		r.send <- v
	}()
}
