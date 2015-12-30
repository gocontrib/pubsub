package test

import (
	"os"
	"testing"

	"github.com/gocontrib/pubsub/redis"
)

func TestRedis_Basic(t *testing.T) {
	url := os.Getenv("REDIS_URL")
	if len(url) == 0 {
		url = "tcp://127.0.0.1:6379/11"
	}
	hub, err := redis.Open(url)
	ok(t, "Open", err)
	verifyBasicAPI(t, hub)
}
