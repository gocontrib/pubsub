package redis

import (
	"os"

	"github.com/gocontrib/pubsub"
	log "github.com/sirupsen/logrus"
	"github.com/soveran/redisurl"
)

func init() {
	pubsub.RegisterDriver(&driver{}, "redis")
}

type driver struct{}

func (d *driver) Create(config pubsub.HubConfig) (pubsub.Hub, error) {
	log.Info("connecting to redis pubsub")
	return Open(config.GetString("url", ""))
}

// Open creates pubsub hub connected to redis server.
func Open(URL ...string) (pubsub.Hub, error) {
	redisURL := getRedisURL(URL...)
	conn, err := redisurl.ConnectToURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &hub{
		conn:     conn,
		redisURL: redisURL,
		subs:     make(map[*sub]struct{}),
	}, nil
}

func getRedisURL(URL ...string) string {
	if len(URL) == 1 && len(URL[0]) > 0 {
		return URL[0]
	}
	return os.Getenv("REDIS_URL")
}
