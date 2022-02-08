package registry_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/registry"
	"github.com/platform-edn/courier/pkg/registry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNodeRegistry_RegisterNodes(t *testing.T) {
	var badEvent registry.NodeEventType = 4

	tests := map[string]struct {
		event registry.NodeEvent
	}{
		"adds node to registry": {
			event: registry.NewNodeEvent(
				registry.RemovePointers(
					registry.CreateTestNodes(1, &registry.TestNodeOptions{}),
				)[0],
				registry.Add,
			),
		},
		"removes node from registry on remove": {
			event: registry.NewNodeEvent(
				registry.RemovePointers(
					registry.CreateTestNodes(1, &registry.TestNodeOptions{}),
				)[0],
				registry.Remove,
			),
		},
		"removes node from registry on fail": {
			event: registry.NewNodeEvent(
				registry.RemovePointers(
					registry.CreateTestNodes(1, &registry.TestNodeOptions{}),
				)[0],
				registry.Failed,
			),
		},
		"fails with bad event": {
			event: registry.NewNodeEvent(
				registry.RemovePointers(
					registry.CreateTestNodes(1, &registry.TestNodeOptions{}),
				)[0],
				badEvent,
			),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ctx, cancel := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			wg.Add(1)
			errChan := make(chan error, 1)

			nodeReg := registry.NewNodeRegistry()
			subLister := new(mocks.SubscriberLister)
			nodeMapper := new(mocks.NodeMapper)
			nodeReg.SubscriberLister = subLister
			nodeReg.NodeMapper = nodeMapper

			subLister.On("CloseListeners").Return()

			switch test.event.Event {
			case registry.Add:
				nodeMapper.On("AddNode", test.event.Node).Return()
				subLister.On(("ForwardEvent"), test.event).Return()
			case registry.Remove:
				nodeMapper.On("RemoveNode", mock.Anything).Return()
				subLister.On(("ForwardEvent"), test.event).Return()
			case registry.Failed:
				nodeMapper.On("RemoveNode", mock.Anything).Return()
				subLister.On(("ForwardEvent"), test.event).Return()
			}

			done := make(chan struct{})
			go func() {
				nodeReg.RegisterNodes(ctx, wg, errChan)

				nodeMapper.AssertExpectations(t)
				subLister.AssertExpectations(t)

				switch test.event.Event {
				case registry.Add:
					nodeMapper.AssertNumberOfCalls(t, "AddNode", 1)
					subLister.AssertNumberOfCalls(t, "ForwardEvent", 1)
				case registry.Remove:
					nodeMapper.AssertNumberOfCalls(t, "RemoveNode", 1)
					subLister.AssertNumberOfCalls(t, "ForwardEvent", 1)
				case registry.Failed:
					nodeMapper.AssertNumberOfCalls(t, "RemoveNode", 1)
					subLister.AssertNumberOfCalls(t, "ForwardEvent", 1)
				}

				subLister.AssertNumberOfCalls(t, "CloseListeners", 1)

				close(errChan)
				errs := []error{}
				for err := range errChan {
					errs = append(errs, err)
				}

				if test.event.Event == badEvent {
					assert.Len(errs, 1)
				} else {
					assert.Len(errs, 0)
				}

				close(done)
			}()

			nodeReg.EventInChannel() <- test.event
			cancel()

			select {
			case <-done:
				break
			case <-time.After(time.Second * 3):
				t.Fatal("did not finish in time")
			}

		})
	}
}
