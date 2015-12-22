package nsq

import (
	"fmt"
	"strings"

	"github.com/drone/config"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nsqio/go-nsq"
)

var (
	// TODO allow multiple addresses
	nsqdAddr       = config.String("pubsub-nsqd", "127.0.0.1:4150")
	nsqLookupdAddr = config.String("pubsub-nsqlookupd", "127.0.0.1:4160")
	nsqMaxInFlight = config.Int("pubsub-nsq-maxinflight", 1000)
)

func init() {
	pubsub.RegisterDriver(&driver{}, "nsq", "nsqio")
}

type driver struct{}

func (d *driver) Create() (pubsub.Hub, error) {
	log.Info("connecting to nsq pubsub")

	var addr = *nsqdAddr
	var producer, err = nsq.NewProducer(addr, nsqConfig())
	if err != nil {
		return nil, err
	}

	return &hub{addr, *nsqLookupdAddr, producer}, nil
}

func nsqConfig() *nsq.Config {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("nsq_pubsub/%s go-nsq/%s", "0.0.1", nsq.VERSION)
	cfg.MaxInFlight = *nsqMaxInFlight
	return cfg
}

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

// Subscription channel.
type sub struct {
	consumers []*nsq.Consumer
	closed    chan bool
	send      chan interface{}
}

func (s *sub) Read() <-chan interface{} {
	return s.send
}

func (s *sub) Close() error {
	go func() {
		s.closed <- true
	}()
	go func() {
		for _, c := range s.consumers {
			c.Stop()
		}
	}()
	return nil
}

func (s *sub) CloseNotify() <-chan bool {
	return s.closed
}

func (s *sub) HandleMessage(msg *nsq.Message) error {
	go func() {
		v, err := pubsub.Unmarshal(msg.Body)
		if err != nil {
			return
		}
		s.send <- v
	}()
	return nil
}

func escapeChannelName(name string) string {
	return strings.Replace(name, "/", "-", -1)
}
