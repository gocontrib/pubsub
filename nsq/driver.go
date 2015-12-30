package nsq

import (
	"fmt"

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
