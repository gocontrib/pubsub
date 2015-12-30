package pubsub

import (
	"sync"
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
	cn, ok := hub.channels[name]
	if ok {
		return cn
	}
	hub.Lock()
	defer hub.Unlock()
	cn = makeChannel(hub, name)
	hub.channels[name] = cn
	go cn.start()
	return cn
}

// Removes given channel, called by Channel.Close.
func (hub *hub) remove(cn *channel) {
	cn, ok := hub.channels[cn.name]
	if !ok {
		return
	}
	hub.Lock()
	defer hub.Unlock()
	delete(hub.channels, cn.name)
	return
}
