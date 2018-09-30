package nats

import (
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/nats-io/nats"
)

func init() {
	pubsub.RegisterDriver(&driver{}, "nats", "natsio")
}

type driver struct{}

func (d *driver) Create(config pubsub.HubConfig) (pubsub.Hub, error) {

	url, ok := config["url"].(string)
	if ok {
		return Open(url)
	}
	return Open("")
}

// Open creates pubsub hub connected to nats server.
func Open(URL ...string) (pubsub.Hub, error) {
	if len(URL) == 0 {
		URL = []string{nats.DefaultURL}
	}

	log.Info("connecting to nats hub: %v", URL)

	conn, err := nats.Connect(URL[0])
	if err != nil {
		return nil, err
	}

	return &hub{
		conn: conn,
		subs: make(map[*sub]struct{}),
	}, nil
}
