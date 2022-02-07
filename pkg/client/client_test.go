package client_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/client"
	"github.com/platform-edn/courier/pkg/client/mocks"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_ListenForResponseInfo(t *testing.T) {
	tests := map[string]struct {
		ids []string
	}{
		"pushes all ResponseInfo in to the ResponseMapper": {
			ids: []string{"id1", "id2", "id3"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			failedNodes := make(chan registry.NodeEvent)
			defer close(failedNodes)
			responses := make(chan messaging.ResponseInfo)
			defer close(responses)
			messager := client.NewMessagingClient(failedNodes, client.WithInsecure())
			respMapper := new(mocks.ResponseMapper)
			messager.ResponseMapper = respMapper
			ctx, cancel := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			wg.Add(1)
			defer cancel()

			respMapper.On("PushResponse", mock.AnythingOfType("messaging.ResponseInfo")).Return()

			go messager.ListenForResponseInfo(ctx, wg, responses)

			done := make(chan struct{})
			go func() {
				for _, id := range test.ids {
					response := messaging.ResponseInfo{
						MessageId: id,
						NodeId:    "nodeId",
					}
					responses <- response
				}

				cancel()
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				respMapper.AssertExpectations(t)
				respMapper.AssertNumberOfCalls(t, "PushResponse", len(test.ids))
			case <-time.After(time.Second * 3):
				t.Fatal("did not return in time")
			}
		})
	}
}

func TestClient_ListenForNodeEvents(t *testing.T) {
	tests := map[string]struct {
		event registry.NodeEventType
		err   error
	}{
		"adds a node to the client node map": {
			event: registry.Add,
			err:   nil,
		},
		"removes a node from the client node map": {
			event: registry.Remove,
			err:   nil,
		},
		"sends an error through the err channel": {
			event: registry.Failed,
			err:   &client.UnknownNodeEventError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			events := make(chan registry.NodeEvent)
			failures := make(chan registry.NodeEvent)
			errs := make(chan error, 1)
			messager := client.NewMessagingClient(failures, client.WithInsecure())
			ctx, cancel := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}

			defer close(events)
			defer close(failures)
			defer cancel()
			wg.Add(1)

			nodeMapper := new(mocks.ClientNodeMapper)
			subMapper := new(mocks.SubMapper)
			messager.ClientNodeMapper = nodeMapper
			messager.SubMapper = subMapper

			switch test.event {
			case registry.Add:
				nodeMapper.On("AddClientNode", mock.AnythingOfType("client.ClientNode")).Return()
			case registry.Remove:
				nodeMapper.On("RemoveClientNode", mock.AnythingOfType("string")).Return()
				subMapper.On("RemoveSubscriber", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()
			default:
			}

			go messager.ListenForNodeEvents(ctx, wg, events, errs, "nodeId")

			done := make(chan struct{})
			go func() {
				node := registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{
					SubscribedSubjects: []string{"sub"},
				}))[0]
				event := registry.NewNodeEvent(node, test.event)
				events <- *event

				cancel()
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				switch test.event {
				case registry.Add:
					nodeMapper.AssertExpectations(t)
					nodeMapper.AssertNumberOfCalls(t, "AddClientNode", 1)
				case registry.Remove:
					nodeMapper.AssertExpectations(t)
					subMapper.AssertExpectations(t)
					nodeMapper.AssertNumberOfCalls(t, "RemoveClientNode", 1)
					subMapper.AssertNumberOfCalls(t, "RemoveSubscriber", 1)
				default:
					close(errs)
					assert.Error(test.err)
					nodeMapper.AssertNumberOfCalls(t, "RemoveClientNode", 0)
					subMapper.AssertNumberOfCalls(t, "RemoveSubscriber", 0)
					nodeMapper.AssertNumberOfCalls(t, "AddClientNode", 0)
					for err := range errs {
						errorType := test.err
						assert.ErrorAs(err, &errorType)
					}

					return
				}
			case <-time.After(time.Second * 3):
				t.Fatal("did not return in time")
			}

			close(errs)
			errList := []error{}
			for err := range errs {
				errList = append(errList, err)
			}

			assert.NoError(test.err)
			assert.Len(errList, 0)
		})
	}
}

func TestClient_Publish(t *testing.T) {
	tests := map[string]struct {
		messageType    messaging.MessageType
		contextTimeout time.Duration
		generateErr    error
		err            error
	}{
		"sends messages to subscribed nodes adn returns": {
			messageType:    messaging.PubMessage,
			contextTimeout: time.Second * 3,
			err:            nil,
			generateErr:    nil,
		},
		"returns error when given in correct message type": {
			messageType:    messaging.ReqMessage,
			contextTimeout: time.Second * 3,
			err:            &client.BadMessageTypeError{},
			generateErr:    nil,
		},
		"returns error when generating ids fails": {
			messageType:    messaging.PubMessage,
			contextTimeout: time.Second * 3,
			err:            nil,
			generateErr:    errors.New("bad!"),
		},
		"returns error when context finishes first": {
			messageType:    messaging.PubMessage,
			contextTimeout: time.Microsecond,
			err:            &client.ContextDoneUnsentMessageError{},
			generateErr:    nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			failed := make(chan registry.NodeEvent)
			failedNodes := make(chan registry.Node)
			close(failedNodes)
			ctx, cancel := context.WithTimeout(context.Background(), test.contextTimeout)
			msg := messaging.Message{
				Id:      "id",
				Type:    test.messageType,
				Subject: "test",
				Content: []byte("test"),
			}

			defer close(failed)
			defer cancel()

			messager := client.NewMessagingClient(failed, client.WithInsecure())
			subMapper := new(mocks.SubMapper)
			clientMapper := new(mocks.ClientNodeMapper)
			messager.ClientNodeMapper = clientMapper
			messager.SubMapper = subMapper

			subMapper.On("GenerateIdsBySubject", msg.Subject).Return(nil, test.generateErr)
			clientMapper.On("FanClientNodeMessaging", ctx, msg, mock.Anything).Return(failedNodes)

			err := messager.Publish(ctx, msg)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			} else if test.generateErr != nil {
				errorType := test.generateErr
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			subMapper.AssertExpectations(t)
			clientMapper.AssertExpectations(t)

			subMapper.AssertNumberOfCalls(t, "GenerateIdsBySubject", 1)
			clientMapper.AssertNumberOfCalls(t, "FanClientNodeMessaging", 1)
		})
	}
}

func TestClient_Response(t *testing.T) {
	tests := map[string]struct {
		messageType    messaging.MessageType
		contextTimeout time.Duration
		generateErr    error
		err            error
	}{
		"sends messages to subscribed nodes and returns": {
			messageType:    messaging.RespMessage,
			contextTimeout: time.Second * 3,
			err:            nil,
			generateErr:    nil,
		},
		"returns error when given in correct message type": {
			messageType:    messaging.PubMessage,
			contextTimeout: time.Second * 3,
			err:            &client.BadMessageTypeError{},
			generateErr:    nil,
		},
		"returns error when generating ids fails": {
			messageType:    messaging.RespMessage,
			contextTimeout: time.Second * 3,
			err:            nil,
			generateErr:    errors.New("bad!"),
		},
		"returns error when context finishes first": {
			messageType:    messaging.RespMessage,
			contextTimeout: time.Microsecond,
			err:            &client.ContextDoneUnsentMessageError{},
			generateErr:    nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			failed := make(chan registry.NodeEvent)
			failedNodes := make(chan registry.Node)
			close(failedNodes)
			ctx, cancel := context.WithTimeout(context.Background(), test.contextTimeout)
			msg := messaging.Message{
				Id:      "id",
				Type:    test.messageType,
				Subject: "test",
				Content: []byte("test"),
			}

			defer close(failed)
			defer cancel()

			messager := client.NewMessagingClient(failed, client.WithInsecure())
			respMapper := new(mocks.ResponseMapper)
			clientMapper := new(mocks.ClientNodeMapper)
			messager.ClientNodeMapper = clientMapper
			messager.ResponseMapper = respMapper

			respMapper.On("GenerateIdsByMessage", msg.Id).Return(nil, test.generateErr)
			clientMapper.On("FanClientNodeMessaging", ctx, msg, mock.Anything).Return(failedNodes)

			err := messager.Response(ctx, msg)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			} else if test.generateErr != nil {
				errorType := test.generateErr
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			respMapper.AssertExpectations(t)
			clientMapper.AssertExpectations(t)

			respMapper.AssertNumberOfCalls(t, "GenerateIdsByMessage", 1)
			clientMapper.AssertNumberOfCalls(t, "FanClientNodeMessaging", 1)
		})
	}
}

func TestClient_Request(t *testing.T) {
	tests := map[string]struct {
		messageType    messaging.MessageType
		contextTimeout time.Duration
		generateErr    error
		err            error
	}{
		"sends messages to subscribed nodes adn returns": {
			messageType:    messaging.ReqMessage,
			contextTimeout: time.Second * 3,
			err:            nil,
			generateErr:    nil,
		},
		"returns error when given in correct message type": {
			messageType:    messaging.PubMessage,
			contextTimeout: time.Second * 3,
			err:            &client.BadMessageTypeError{},
			generateErr:    nil,
		},
		"returns error when generating ids fails": {
			messageType:    messaging.ReqMessage,
			contextTimeout: time.Second * 3,
			err:            nil,
			generateErr:    errors.New("bad!"),
		},
		"returns error when context finishes first": {
			messageType:    messaging.ReqMessage,
			contextTimeout: time.Microsecond,
			err:            &client.ContextDoneUnsentMessageError{},
			generateErr:    nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			failed := make(chan registry.NodeEvent)
			failedNodes := make(chan registry.Node)
			close(failedNodes)
			ctx, cancel := context.WithTimeout(context.Background(), test.contextTimeout)
			msg := messaging.Message{
				Id:      "id",
				Type:    test.messageType,
				Subject: "test",
				Content: []byte("test"),
			}

			defer close(failed)
			defer cancel()

			messager := client.NewMessagingClient(failed, client.WithInsecure())
			subMapper := new(mocks.SubMapper)
			clientMapper := new(mocks.ClientNodeMapper)
			messager.ClientNodeMapper = clientMapper
			messager.SubMapper = subMapper

			subMapper.On("GenerateIdsBySubject", msg.Subject).Return(nil, test.generateErr)
			clientMapper.On("FanClientNodeMessaging", ctx, msg, mock.Anything).Return(failedNodes)

			err := messager.Request(ctx, msg)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			} else if test.generateErr != nil {
				errorType := test.generateErr
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			subMapper.AssertExpectations(t)
			clientMapper.AssertExpectations(t)

			subMapper.AssertNumberOfCalls(t, "GenerateIdsBySubject", 1)
			clientMapper.AssertNumberOfCalls(t, "FanClientNodeMessaging", 1)
		})
	}
}

func TestClient_ForwardFailedConnections(t *testing.T) {
	tests := map[string]struct {
		ids []string
	}{
		"all failed nodes passed in are passed out": {
			ids: []string{"id1", "id2", "id3"},
		},
		"doesn't fail when nothing is passed through": {
			ids: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			failed := make(chan registry.NodeEvent, len(test.ids))
			messager := client.NewMessagingClient(failed, client.WithInsecure())
			subMapper := new(mocks.SubMapper)
			clientMapper := new(mocks.ClientNodeMapper)
			messager.ClientNodeMapper = clientMapper
			messager.SubMapper = subMapper

			clientMapper.On("RemoveClientNode", mock.AnythingOfType("string")).Return()
			subMapper.On("RemoveSubscriber", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()

			failedNodes := make(chan registry.Node, len(test.ids))

			for _, id := range test.ids {
				node := registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{
					SubscribedSubjects: []string{"sub"},
					Id:                 id,
				}))[0]

				failedNodes <- node
			}

			close(failedNodes)
			done := messager.ForwardFailedConnections(failedNodes)

			select {
			case <-done:
				break
			case <-time.After(time.Second * 3):
				t.Fatal("did not finish in time")
			}

			if len(test.ids) > 0 {
				clientMapper.AssertExpectations(t)
				subMapper.AssertExpectations(t)
				clientMapper.AssertNumberOfCalls(t, "RemoveClientNode", len(test.ids))
				subMapper.AssertNumberOfCalls(t, "RemoveSubscriber", len(test.ids))
			}

			close(failed)

			events := []registry.NodeEvent{}
			for event := range failed {
				events = append(events, event)
			}

			assert.Len(events, len(test.ids))

		})
	}
}
