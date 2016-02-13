package nsq

import (
	"fmt"

	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nsqio/go-nsq"
)

type nsqConfig struct {
	config pubsub.HubConfig
}

// TODO allow multiple addresses
func (c nsqConfig) nodeAddr() string {
	return c.config.GetString("nsqd", "127.0.0.1:4150")
}

func (c nsqConfig) lookupdAddr() string {
	return c.config.GetString("nsqlookupd", "127.0.0.1:4160")
}

func (c nsqConfig) maxInFlight() int {
	return c.config.GetInt("maxinflight", 1000)
}

func init() {
	pubsub.RegisterDriver(&driver{}, "nsq", "nsqio")
}

type driver struct{}

func (d *driver) Create(config pubsub.HubConfig) (pubsub.Hub, error) {
	log.Info("connecting to nsq pubsub")

	cfg := nsqConfig{config}

	addr := cfg.nodeAddr()
	producer, err := nsq.NewProducer(addr, makeConfig(cfg))
	if err != nil {
		return nil, err
	}

	return &hub{cfg, producer}, nil
}

func makeConfig(config nsqConfig) *nsq.Config {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("nsq_pubsub/%s go-nsq/%s", "0.0.1", nsq.VERSION)
	cfg.MaxInFlight = config.maxInFlight()
	return cfg
}
