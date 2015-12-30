package nsq

import (
	"strings"

	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nsqio/go-nsq"
)

// NSQ pubsub hub
type hub struct {
	nodeAddr    string
	lookupdAddr string
	producer    *nsq.Producer
}

func (h *hub) Publish(channels []string, msg interface{}) {
	go func() {
		var body, err = pubsub.Marshal(msg)
		if err != nil {
			return
		}
		for _, name := range channels {
			h.producer.Publish(escapeChannelName(name), body)
		}
	}()
}

func (h *hub) Subscribe(channels []string) (pubsub.Channel, error) {
	s := &sub{
		closed: make(chan bool),
		send:   make(chan interface{}),
	}
	for _, name := range channels {
		var c, err = h.makeConsumer(escapeChannelName(name), s)
		if err != nil {
			return nil, err
		}
		s.consumers = append(s.consumers, c)
	}
	return s, nil
}

func (h *hub) makeConsumer(topic string, handler nsq.Handler) (*nsq.Consumer, error) {
	// TODO fix channel name
	var c, err = nsq.NewConsumer(topic, "hub", nsqConfig())
	if err != nil {
		return nil, err
	}

	c.AddHandler(handler)

	err = c.ConnectToNSQLookupd(h.lookupdAddr)
	if err != nil {
		log.Errorf("cannot connect to nsqlookupd at %s: %v", h.lookupdAddr, err)
	} else {
		return c, nil
	}

	err = c.ConnectToNSQD(h.nodeAddr)
	if err != nil {
		log.Errorf("cannot connect to nsqd at %s: %v", h.nodeAddr, err)
		return nil, err
	}

	return c, nil
}

func (h *hub) Close() error {
	h.producer.Stop()
	return nil
}

func escapeChannelName(name string) string {
	return strings.Replace(name, "/", "-", -1)
}
