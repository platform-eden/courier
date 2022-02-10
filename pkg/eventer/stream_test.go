package eventer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/platform-edn/courier/pkg/eventer"
	"github.com/platform-edn/courier/pkg/proto"
	"github.com/platform-edn/courier/pkg/proto/mocks"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventStream_StreamEvents(t *testing.T) {
	subErr := errors.New("sub error")
	streamErr := errors.New("bad stream")

	tests := map[string]struct {
		subscribeErr error
		streamErr    error
		assertedErr  error
	}{
		"should error from subscribe": {
			subscribeErr: subErr,
			streamErr:    streamErr,
			assertedErr:  subErr,
		},
		"should error from stream": {
			subscribeErr: nil,
			streamErr:    streamErr,
			assertedErr:  streamErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()
			out := make(chan registry.NodeEvent, 1)
			client := new(mocks.DiscoverClient)
			evClient := new(mocks.Discover_SubscribeToNodeEventsClient)
			subjects := []string{"sub"}
			nodeId := "test"
			nodeAddres := "test"
			nodePort := "8000"
			node := &proto.Node{
				Id:                  nodeId,
				Address:             nodeAddres,
				Port:                nodePort,
				BroadcastedSubjects: subjects,
				SubscribedSubjects:  subjects,
			}

			stream := eventer.NewEventStream(out, client)

			client.On("SubscribeToNodeEvents", ctx, mock.Anything).Return(evClient, test.subscribeErr)

			if test.subscribeErr == nil {
				evClient.On("Recv").Return(&proto.EventStreamResponse{
					Event: &proto.NodeEvent{
						EventType: proto.NodeEventType_Added,
						Node:      node,
					},
				}, nil).Once()
				evClient.On("Recv").Return(nil, test.streamErr)
			}

			err := stream.StreamEvents(ctx, subjects, subjects)
			errorType := test.assertedErr
			assert.ErrorAs(err, &errorType)

			client.AssertExpectations(t)
			evClient.AssertExpectations(t)

			if test.subscribeErr == nil {
				evClient.AssertNumberOfCalls(t, "Recv", 2)
				close(out)

				for event := range out {
					assert.Equal(event.Address, nodeAddres)
					assert.Equal(event.Id, nodeId)
					assert.Equal(event.Port, nodePort)
					assert.EqualValues(event.BroadcastedSubjects, subjects)
					assert.EqualValues(event.SubscribedSubjects, subjects)
					assert.Equal(event.Event, registry.Add)
				}
				return
			}

			close(out)
		})
	}
}
