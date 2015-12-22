package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocontrib/pubsub"
)

func OK(t *testing.T, op string, err error) {
	if err != nil {
		t.Errorf("%s failed: %+v", op, err)
		t.FailNow()
	}
}

func MustReceive(t *testing.T, event <-chan bool) {
	select {
	case <-event:
	case <-time.After(time.Second * 1):
		t.Error("timeout")
		t.FailNow()
	}
}

func VerifyBasicAPI(t *testing.T, hub pubsub.Hub) {
	s, err := hub.Subscribe([]string{"test"})
	OK(t, "Subscribe", err)

	var msg interface{}
	msgReceived := make(chan bool)
	closeReceived := make(chan bool)

	go func() {
		for {
			select {
			case m := <-s.Read():
				fmt.Println("msg received")
				msg = m
				msgReceived <- true
			case <-s.CloseNotify():
				fmt.Println("close received")
				closeReceived <- true
				return
			}
		}
	}()

	hub.Publish([]string{"test"}, "test")

	MustReceive(t, msgReceived)

	hub.Close()

	MustReceive(t, closeReceived)
}

func TestHub_Basic(t *testing.T) {
	VerifyBasicAPI(t, pubsub.NewHub())
}
