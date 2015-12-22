package test

import (
	"testing"

	"github.com/gocontrib/pubsub"
)

func TestHub_Basic(t *testing.T) {
	verifyBasicAPI(t, pubsub.NewHub())
}
