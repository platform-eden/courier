package courier

import (
	"testing"
	"time"
)

func TestNewQueueHandler(t *testing.T) {
	_, err := NewQueueHandler()
	if err != nil {
		t.Fatalf("could not create queue: %s", err)
	}

}

func TestQueueHandler_SubscribeAndPush(t *testing.T) {
	qh, err := NewQueueHandler()
	if err != nil {
		t.Fatalf("could not create queue: %s", err)
	}

	pop := qh.Subscribe("test")
	push := qh.PushChannel()

	m := NewPubMessage("test", []byte("test"))

	push <- m

	select {
	case msg := <-pop:
		if msg.Subject != "test" {
			t.Fatalf("expected subject to be test but got %s", msg.Subject)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("timeout waitng on subscribe channel")
	}
}
