package nats

import (
	"github.com/drone/config"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nats-io/nats"
)

var (
	natsURL = config.String("pubsub-nats", "")
)

func init() {
	pubsub.RegisterDriver(&driver{}, "nats", "natsio")
}

type driver struct{}

func (d *driver) Create() (pubsub.Hub, error) {
	log.Info("connecting to nats hub")
	return Open()
}

// Open creates pubsub hub connected to nats server.
func Open() (pubsub.Hub, error) {
	var url = *natsURL
	if len(url) == 0 {
		url = nats.DefaultURL
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &hub{
		conn: conn,
		subs: make(map[*sub]struct{}),
	}, nil
}
