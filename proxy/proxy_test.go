package proxy

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/message"
)

func TestMessageProxy_Subscribe(t *testing.T) {
	push := make(chan message.Message)
	defer close(push)
	quit := make(chan bool)
	defer close(quit)

	p := NewMessageProxy(push, quit)

	sub := p.Subscribe("test")

	go func() {
		m1 := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
		sub <- m1
	}()

	select {
	case msg := <-sub:
		if msg.Subject != "test" {
			t.Fatalf("expected subject to be test but got %s", msg.Subject)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("timeout waitng on subscribe channel")
	}

	quit <- true
}

func TestSend(t *testing.T) {
	push := make(chan message.Message)
	defer close(push)
	quit := make(chan bool)
	defer close(quit)

	p := NewMessageProxy(push, quit)

	m1 := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	err := send(m1, p.Subscriptions)
	if err == nil {
		t.Fatalf("expected sending on nonexistent subject to fail but it passed")
	}

	sub := p.Subscribe("test")

	err = send(m1, p.Subscriptions)
	if err != nil {
		t.Fatalf("expected send to pass but it failed: %s", err)
	}

	select {
	case msg := <-sub:
		if msg.Subject != "test" {
			t.Fatalf("expected subject to be test but got %s", msg.Subject)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("timeout waitng on subscribe channel")
	}

	quit <- true
}

func TestMessageProxy_Start(t *testing.T) {
	push := make(chan message.Message)
	defer close(push)
	quit := make(chan bool)
	defer close(quit)

	p := NewMessageProxy(push, quit)

	sub := p.Subscribe("test")

	pchan := p.PushChannel

	m1 := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	m2 := message.NewPubMessage(uuid.NewString(), "test1", []byte("test"))

	pchan <- m1
	pchan <- m2

	select {
	case msg := <-sub:
		if msg.Subject != "test" {
			t.Fatalf("expected subject to be test but got %s", msg.Subject)
		}
	case <-time.After(time.Second * 1):
		t.Fatal("timeout waitng on subscribe channel")
	}

	quit <- true
}
