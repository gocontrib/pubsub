package pubsub

import (
	"runtime/debug"
	"sync"

	"github.com/gocontrib/log"
)

// NewHub creates new in-process pubsub hub.
func NewHub() Hub {
	return &hub{
		channels: make(map[string]*channel),
	}
}

// Hub of pubsub channels.
type hub struct {
	sync.Mutex
	channels map[string]*channel
}

func (hub *hub) Close() error {
	for _, c := range hub.channels {
		c.Close()
	}
	return nil
}

// Publish data to given channel.
func (hub *hub) Publish(channels []string, msg interface{}) {
	for _, name := range channels {
		var cn = hub.getChannel(name)
		cn.Publish(msg)
	}
}

// Subscribe adds new receiver of events for given channel.
func (hub *hub) Subscribe(channels []string) (Channel, error) {
	var chans []*channel
	for _, name := range channels {
		chans = append(chans, hub.getChannel(name))
	}
	var sub = makeSub(chans)
	for _, cn := range chans {
		cn.Subscribe(sub)
	}
	return sub, nil
}

// GetChannel gets or creates new pubsub channel.
func (hub *hub) getChannel(name string) *channel {
	hub.Lock()
	defer hub.Unlock()
	cn, ok := hub.channels[name]
	if ok {
		return cn
	}
	cn = makeChannel(hub, name)
	hub.channels[name] = cn
	go cn.start()
	return cn
}

// Removes given channel, called by Channel.Close.
func (hub *hub) remove(cn *channel) {
	hub.Lock()
	defer hub.Unlock()
	cn, ok := hub.channels[cn.name]
	if !ok {
		return
	}
	delete(hub.channels, cn.name)
	return
}

// Channel manages topic subscriptions.
type channel struct {
	hub         *hub
	name        string
	closed      chan bool
	broadcast   chan interface{}
	subscribe   chan *sub
	unsubscribe chan *sub
	subs        map[*sub]struct{}
}

func makeChannel(hub *hub, name string) *channel {
	return &channel{
		hub:         hub,
		name:        name,
		closed:      make(chan bool),
		broadcast:   make(chan interface{}),
		subscribe:   make(chan *sub),
		unsubscribe: make(chan *sub),
		subs:        make(map[*sub]struct{}),
	}
}

// Publish data to all subscribers.
func (c *channel) Publish(data interface{}) {
	go func() { c.broadcast <- data }()
}

// Subscribe adds new receiver.
func (c *channel) Subscribe(r *sub) {
	go func() { c.subscribe <- r }()
}

// Close channel.
func (c *channel) Close() {
	go func() { c.closed <- true }()
}

func (c *channel) start() {
	if err := recover(); err != nil {
		log.Errorf("recovered from panic: %+v", err)
		debug.PrintStack()
	}

	for {
		select {

		case sub := <-c.subscribe:
			c.subs[sub] = struct{}{}

		case sub := <-c.unsubscribe:
			delete(c.subs, sub)
			close(sub.send)

		case msg := <-c.broadcast:
			for sub := range c.subs {
				sub.send <- msg
			}

		case <-c.closed:
			c.stop()
			return
		}
	}
}

func (c *channel) stop() {
	for s := range c.subs {
		s.Close()
	}
	c.hub.remove(c)
}

// Subscription to multiple hub channels.
type sub struct {
	channels []*channel
	closed   chan bool
	send     chan interface{}
}

func makeSub(channels []*channel) *sub {
	return &sub{
		channels: channels,
		closed:   make(chan bool),
		send:     make(chan interface{}),
	}
}

// Read returns channel of receiver events.
func (s *sub) Read() <-chan interface{} {
	return s.send
}

// Close removes subscriber from channel.
func (s *sub) Close() error {
	go func() {
		for _, c := range s.channels {
			c.unsubscribe <- s
		}
	}()
	go func() {
		s.closed <- true
		close(s.send)
	}()
	return nil
}

// CloseNotify returns channel to handle close event.
func (s *sub) CloseNotify() <-chan bool {
	return s.closed
}
