package registry_test

import (
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
)

func TestNodeEventSubscribers_SubscribeToEvents(t *testing.T) {
	assert := assert.New(t)
	subscribers := registry.NewNodeEventSubscribers()
	channel := subscribers.SubscribeToEvents()

	assert.Len(subscribers.Channels, 1)
	assert.Contains(subscribers.Channels, channel)
}

func TestNodeEventSubscribers_CloseListeners(t *testing.T) {
	subscribers := registry.NewNodeEventSubscribers()
	channel := subscribers.SubscribeToEvents()

	done := make(chan struct{})
	go func() {
		<-channel

		close(done)
	}()

	subscribers.CloseListeners()

	select {
	case <-done:
	case <-time.After(time.Second * 3):
		t.Fatal("did not finish in time")
	}
}

func TestNodeEventSubscribers_ForwardEvent(t *testing.T) {
	subscribers := registry.NewNodeEventSubscribers()
	channel := subscribers.SubscribeToEvents()
	event := registry.NewNodeEvent(
		registry.RemovePointers(
			registry.CreateTestNodes(1, &registry.TestNodeOptions{
				Id:   "test",
				Port: "8000",
			}),
		)[0],
		registry.Add,
	)

	done := make(chan struct{})
	go func() {
		ev := <-channel

		assert.EqualValues(t, event, ev)
		close(done)
	}()

	subscribers.ForwardEvent(event)

	select {
	case <-done:
	case <-time.After(time.Second * 3):
		t.Fatal("did not finish in time")
	}
}
