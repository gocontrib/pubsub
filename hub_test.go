package pubsub

import (
	"testing"
	"time"
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

func VerifyBasicAPI(t *testing.T, h Hub) {
	s, err := h.Subscribe([]string{"test"})
	OK(t, "Subscribe", err)

	var msg interface{}
	msgReceived := make(chan bool)
	closeReceived := make(chan bool)

	go func() {
		for {
			select {
			case m := <-s.Read():
				msg = m
				msgReceived <- true
			case <-s.CloseNotify():
				closeReceived <- true
				return
			}
		}
	}()

	h.Publish([]string{"test"}, "test")

	MustReceive(t, msgReceived)

	h.Close()

	MustReceive(t, closeReceived)
}

func TestBasic(t *testing.T) {
	VerifyBasicAPI(t, newHub())
}
