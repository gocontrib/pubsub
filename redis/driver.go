package redis

import (
	"os"

	"github.com/drone/config"
	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
	"github.com/soveran/redisurl"
)

var (
	redisURL = config.String("pubsub-redis", "")
)

func init() {
	pubsub.RegisterDriver(&driver{}, "redis")
}

type driver struct{}

func (d *driver) Create() (pubsub.Hub, error) {
	log.Info("connecting to redis pubsub")
	return Open()
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
	if len(*redisURL) == 0 {
		return os.Getenv("REDIS_URL")
	}
	return *redisURL
}
