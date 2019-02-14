package pubsub

import (
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

// MakeHub returns new instance of the pubsub hub.
func MakeHub(config HubConfig) (Hub, error) {
	if config == nil {
		return NewHub(), nil
	}

	driverName := getDriverName(config)
	if len(driverName) == 0 {
		return nil, fmt.Errorf("driver name is not specified")
	}
	d, ok := drivers[driverName]
	if ok {
		h, err := d.Create(config)
		if err != nil {
			log.Errorf("unable to connect to %s pubsub server: %+v", driverName, err)
			return nil, err
		}
		log.Info("connected to %s pubsub", driverName)
		return h, nil
	}

	return nil, fmt.Errorf("unknown driver: %s", driverName)
}

func getDriverName(config HubConfig) string {
	val := strings.ToLower(strings.TrimSpace(config.GetString("driver", "")))
	if len(val) > 0 {
		return val
	}
	return strings.ToLower(strings.TrimSpace(config.GetString("name", "")))
}

func initOne(config HubConfig) error {
	h, err := MakeHub(config)
	if err != nil {
		return err
	}
	hubInstance = h
	return nil
}

func Init(config HubConfig) (err error) {
	for attemt := 0; attemt < 30; attemt = attemt + 1 {
		err = initOne(config)
		if err == nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
	return err
}
