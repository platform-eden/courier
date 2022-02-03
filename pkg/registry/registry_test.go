package registry

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNodeRegistry_RegisterNodes(t *testing.T) {
	type test struct {
		event *NodeEvent
	}

	node := RemovePointers(CreateTestNodes(1, &TestNodeOptions{}))[0]

	tests := []test{
		{
			event: NewNodeEvent(node, Add),
		},
		{
			event: NewNodeEvent(node, Remove),
		},
		{
			event: NewNodeEvent(node, Failed),
		},
	}

	for _, tc := range tests {
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		errChan := make(chan error)
		defer wg.Add(1)
		defer cancel()

		registry := NewNodeRegistry()

		switch tc.event.Event {
		case Remove:
			registry.addNode(tc.event.Node)
		case Failed:
			registry.addNode(tc.event.Node)
		}

		go registry.RegisterNodes(ctx, wg, errChan)

		in := registry.EventInChannel()
		out := registry.SubscribeToEvents()

		in <- *tc.event

		select {
		case e := <-out:
			if e.Event != tc.event.Event {
				t.Fatalf("expected type %s but got %s", tc.event.Event, e.Event)
			}
		case <-time.After(time.Second * 3):
			t.Fatal("didn't set current in time")
		}

		wg.Add(1)
		waitChannel := make(chan struct{})
		go func() {
			cancel()
			wg.Wait()
			close(waitChannel)
		}()

		select {
		case <-waitChannel:
			continue
		case <-time.After(time.Second * 3):
			t.Fatal("didn't complete wait group in time")
		}
	}
}
