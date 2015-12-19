package pubsub

import (
	"strings"

	"github.com/gocontrib/log"
)

// Driver defines interface for pubsub modules
type Driver interface {
	Create() (Hub, error)
}

var (
	drivers = make(map[string]Driver)
)

// RegisterDriver registers driver with given names.
func RegisterDriver(d Driver, knownNames ...string) {
	for _, k := range knownNames {
		drivers[strings.ToLower(k)] = d
	}
	log.Info("registered pubsub driver: %v", knownNames)
}
