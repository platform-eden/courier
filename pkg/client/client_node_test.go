package client_test

import (
	"errors"
	"testing"
	"time"

	"github.com/platform-edn/courier/pkg/client"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/messaging/mocks"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestNewClientNode(t *testing.T) {
	tests := map[string]struct {
		nodeId  string
		port    string
		err     error
		options []client.ClientNodeOption
	}{
		"creates client node": {
			nodeId: "test",
			port:   "8000",
			options: []client.ClientNodeOption{
				client.WithInsecure(),
				client.WithClientRetryOptions(client.ClientRetryOptionsInput{
					MaxAttempts:     1,
					BackOff:         time.Second,
					Jitter:          1,
					PerRetryTimeout: time.Second,
				}),
			},
			err: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			node := registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{
				Port: test.port,
			}))[0]
			clientNode, err := client.NewClientNode(node, test.nodeId, test.options...)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			assert.EqualValues(node, clientNode.Node)
			assert.Equal(test.nodeId, clientNode.CurrentId)
		})
	}
}

func TestClientNode_AttemptMessage(t *testing.T) {
	tests := map[string]struct {
		nodeId      string
		messageType messaging.MessageType
		message     messaging.Message
		err         error
		options     []client.ClientNodeOption
	}{
		"sends publish message": {
			nodeId:      "test",
			messageType: messaging.PubMessage,
			message:     messaging.NewPubMessage("messageId", "testSubject", []byte("testContent")),
			options: []client.ClientNodeOption{
				client.WithInsecure(),
			},
			err: nil,
		},
		"sends request message": {
			nodeId:      "test",
			messageType: messaging.ReqMessage,
			message:     messaging.NewReqMessage("messageId", "testSubject", []byte("testContent")),
			options: []client.ClientNodeOption{
				client.WithInsecure(),
			},
			err: nil,
		},
		"sends response message": {
			nodeId:      "test",
			messageType: messaging.RespMessage,
			message:     messaging.NewRespMessage("messageId", "testSubject", []byte("testContent")),
			options: []client.ClientNodeOption{
				client.WithInsecure(),
			},
			err: nil,
		},
		"returns an error when sending a publish message": {
			nodeId:      "test",
			messageType: messaging.PubMessage,
			message:     messaging.NewPubMessage("messageId", "testSubject", []byte("testContent")),
			options: []client.ClientNodeOption{
				client.WithInsecure(),
			},
			err: errors.New("bad!"),
		},
		"returns an error when sending a request message": {
			nodeId:      "test",
			messageType: messaging.ReqMessage,
			message:     messaging.NewReqMessage("messageId", "testSubject", []byte("testContent")),
			options: []client.ClientNodeOption{
				client.WithInsecure(),
			},
			err: errors.New("bad!"),
		},
		"returns an error when sending a response message": {
			nodeId:      "test",
			messageType: messaging.RespMessage,
			message:     messaging.NewRespMessage("messageId", "testSubject", []byte("testContent")),
			options: []client.ClientNodeOption{
				client.WithInsecure(),
			},
			err: errors.New("bad!"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			node := registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{}))[0]
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			cli := new(mocks.MessageClient)

			switch test.messageType {
			case messaging.PubMessage:
				cli.On("PublishMessage", ctx, mock.Anything, mock.Anything).Return(&messaging.PublishMessageResponse{}, test.err)
			case messaging.RespMessage:
				cli.On("ResponseMessage", ctx, mock.Anything, mock.Anything).Return(&messaging.ResponseMessageResponse{}, test.err)
			case messaging.ReqMessage:
				cli.On("RequestMessage", ctx, mock.Anything, mock.Anything).Return(&messaging.RequestMessageResponse{}, test.err)
			}

			clientNode, err := client.NewClientNode(node, test.nodeId, test.options...)
			assert.NoError(err)
			clientNode.MessageClient = cli

			err = clientNode.AttemptMessage(ctx, test.message)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			cli.AssertExpectations(t)
		})
	}
}
