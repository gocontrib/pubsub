package test

import (
	"testing"

	"github.com/gocontrib/pubsub/redis"
)

func TestRedis_Basic(t *testing.T) {
	hub, err := redis.Open("tcp://127.0.0.1:6379/11")
	ok(t, "Open", err)
	verifyBasicAPI(t, hub)
}
