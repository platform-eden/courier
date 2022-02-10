package eventer_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/eventer"
	"github.com/platform-edn/courier/pkg/eventer/mocks"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewOperatorClient(t *testing.T) {
	tests := map[string]struct {
		channel     chan registry.NodeEvent
		address     string
		port        string
		subscribed  []string
		broadcasted []string
		options     []grpc.DialOption
		err         error
	}{
		"creates OperatorClient": {
			channel:     make(chan registry.NodeEvent),
			address:     "test",
			port:        "8000",
			subscribed:  []string{"sub"},
			broadcasted: []string{"sub"},
			options:     []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			err:         nil,
		},
		"fails with no address": {
			channel:     make(chan registry.NodeEvent),
			address:     "",
			port:        "8000",
			subscribed:  []string{"sub"},
			broadcasted: []string{"sub"},
			options:     []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			err:         &eventer.MissingOperatorClientParamError{},
		},
		"fails with no port": {
			channel:     make(chan registry.NodeEvent),
			address:     "test",
			port:        "",
			subscribed:  []string{"sub"},
			broadcasted: []string{"sub"},
			options:     []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			err:         &eventer.MissingOperatorClientParamError{},
		},
		"fails with no subjects": {
			channel:     make(chan registry.NodeEvent),
			address:     "test",
			port:        "8000",
			subscribed:  []string{},
			broadcasted: []string{},
			options:     []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			err:         &eventer.MissingOperatorClientParamError{},
		},
		"fails with no dial options": {
			channel:     make(chan registry.NodeEvent),
			address:     "test",
			port:        "8000",
			subscribed:  []string{"sub"},
			broadcasted: []string{"sub"},
			options:     []grpc.DialOption{},
			err:         &eventer.MissingOperatorClientParamError{},
		},
		"fails with no channel": {
			address:     "test",
			port:        "8000",
			subscribed:  []string{"sub"},
			broadcasted: []string{"sub"},
			options:     []grpc.DialOption{},
			err:         &eventer.MissingOperatorClientParamError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client, err := eventer.NewOperatorClient(
				eventer.WithBroadcasted(test.broadcasted...),
				eventer.WithConnectionOptions(test.options...),
				eventer.WithOperatorAddress(test.address),
				eventer.WithOperatorPort(test.port),
				eventer.WithSubscribed(test.subscribed...),
				eventer.WithEventChannel(test.channel),
			)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			assert.Equal(client.OperatorPort, test.port)
			assert.Equal(client.OperatorAddress, test.address)
			assert.EqualValues(client.Subscribed, test.subscribed)
			assert.EqualValues(client.Broadcasted, test.broadcasted)

			close(test.channel)
		})
	}
}

func TestClientOperator_DiscoverNodeEvents(t *testing.T) {
	tests := map[string]struct {
		streamError error
	}{
		"should not error and finish": {
			streamError: nil,
		},
		"should error and finish": {
			streamError: errors.New("bad!"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ctx, cancel := context.WithCancel(context.Background())
			out := make(chan registry.NodeEvent)
			defer close(out)
			defer cancel()
			errs := make(chan error, 1)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			subjects := []string{"sub"}

			client, err := eventer.NewOperatorClient(
				eventer.WithEventChannel(out),
				eventer.WithBroadcasted(subjects...),
				eventer.WithConnectionOptions(grpc.WithTransportCredentials(insecure.NewCredentials())),
				eventer.WithOperatorAddress("test"),
				eventer.WithOperatorPort("8000"),
				eventer.WithSubscribed(subjects...),
			)
			assert.NoError(err)

			streamer := new(mocks.EventStreamer)
			client.EventStreamer = streamer
			streamer.On("StreamEvents", ctx, subjects, subjects).Return(test.streamError).WaitUntil(time.After(time.Second / 2))

			done := make(chan struct{})
			go func() {
				client.DiscoverNodeEvents(ctx, out, errs, wg)
				streamer.AssertExpectations(t)
				streamer.AssertNumberOfCalls(t, "StreamEvents", 1)

				close(errs)
				errList := []error{}
				for err := range errs {
					errList = append(errList, err)
				}

				if test.streamError != nil {
					assert.Len(errList, 1)
					errorType := test.streamError
					assert.ErrorAs(errList[0], &errorType)
				} else {
					assert.Len(errList, 0)
				}

				wg.Wait()
				close(done)
			}()

			<-time.After(time.Second / 4)
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
