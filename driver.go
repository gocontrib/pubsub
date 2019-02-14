package pubsub

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// HubConfig defines input for Driver.Create function.
type HubConfig map[string]interface{}

// GetString property.
func (c HubConfig) GetString(key string, defval string) string {
	val, ok := c[key]
	if ok {
		s, ok := val.(string)
		if ok && len(s) > 0 {
			return s
		}
		// TODO only primitive types
		return fmt.Sprintf("%s", val)
	}
	return defval
}

// GetInt property.
func (c HubConfig) GetInt(key string, defval int) int {
	val, ok := c[key]
	if ok {
		i, ok := val.(int)
		if ok {
			return i
		}
		s := c.GetString(key, "")
		if len(s) > 0 {
			i, err := strconv.Atoi(s)
			if err != nil {
				// TODO handling error on app level
				return defval
			}
			return i
		}
	}
	return defval
}

// Driver defines interface for pubsub modules
type Driver interface {
	Create(config HubConfig) (Hub, error)
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
