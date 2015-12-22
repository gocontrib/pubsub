package test

import (
	"testing"

	"github.com/gocontrib/pubsub/nats"
)

func TestNats_Basic(t *testing.T) {
	hub, err := nats.Open()
	ok(t, "Open", err)
	verifyBasicAPI(t, hub)
}
