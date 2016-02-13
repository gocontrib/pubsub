package nsq

import (
	"fmt"

	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nsqio/go-nsq"
)

type nsqConfig {
	config pubsub.HubConfig
}

// TODO allow multiple addresses
func (c nsqConfig) addr() {
	return c.config.GetString("nsqd", "127.0.0.1:4150")
}

func (c nsqConfig) lookupdAddr() string {
	return c.config.GetString("nsqlookupd", "127.0.0.1:4160")
}

func (c nsqConfig) maxInFlight() int {
	return c.config.Int("maxinflight", 1000)
}

func init() {
	pubsub.RegisterDriver(&driver{}, "nsq", "nsqio")
}

type driver struct{}

func (d *driver) Create(config pubsub.HubConfig) (pubsub.Hub, error) {
	log.Info("connecting to nsq pubsub")

	cfg := nsgConfig{config}

	producer, err := nsq.NewProducer(cfg.addr(), nsqConfig())
	if err != nil {
		return nil, err
	}

	return &hub{addr, cfg.lookupdAddr(), producer}, nil
}

func makeConfig() *nsq.Config {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("nsq_pubsub/%s go-nsq/%s", "0.0.1", nsq.VERSION)
	cfg.MaxInFlight = *nsqMaxInFlight
	return cfg
}
