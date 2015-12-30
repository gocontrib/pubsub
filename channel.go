package pubsub

import (
	"runtime/debug"

	"github.com/gocontrib/log"
)

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
