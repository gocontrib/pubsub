package pubsub

import (
	"errors"
	"strings"

	"github.com/gocontrib/log"
)

var (
	errorNohub  = errors.New("no pubsub engine")
	hubInstance Hub
)

// PUBLIC API

// Cleanup pubsub facilities.
func Cleanup() {
	if hubInstance != nil {
		hubInstance.Close()
		hubInstance = nil
	}
	log.Info("pubsub closed")
}

// Publish message to given channels.
func Publish(channels []string, msg interface{}) error {
	if hubInstance == nil {
		return errorNohub
	}
	log.Debug("publish to %v", channels)
	hubInstance.Publish(channels, msg)
	return nil
}

// Subscribe on given channels.
func Subscribe(channels []string) (Channel, error) {
	if hubInstance == nil {
		return nil, errorNohub
	}
	r, err := hubInstance.Subscribe(channels)
	if err != nil {
		log.Errorf("pubsub subscribe failed: %+v", err)
		return nil, err
	}
	log.Debug("subscibe to %v", channels)
	return r, nil
}

// IMPLEMENTATION

// returns new instance of the pubsub hub.
func makeHub(driverList ...HubConfig) Hub {
	log.Info("starting pubsub")

	for _, config := range driverList {
		name := strings.ToLower(strings.TrimSpace(config.GetString("name", "")))
		if len(name) == 0 {
			continue
		}
		d, ok := drivers[name]
		if ok {
			h, err := d.Create(config)
			if err != nil {
				log.Errorf("unable to connect to %s pubsub server: %+v", name, err)
			}
			log.Info("connected to %s pubsub", name)
			return h
		}
	}

	log.Info("use fallback internal pubsub")
	return NewHub()
}

// Init pubsub hub.
func Init(driverList ...HubConfig) {
	log.Info("init pubsub module")
	hubInstance = makeHub(driverList...)
}
