package client

import (
	"context"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/messaging/proto"
	"github.com/platform-edn/courier/pkg/registry"
	"google.golang.org/grpc"
)

type clientNode struct {
	registry.Node
	connection *grpc.ClientConn
	currentId  string
}

type ClientNodeOptions struct {
	options []grpc.DialOption
}

// CLientNodeOption is a set of options that may be passed as parameters when creating a ClientNode
type ClientNodeOption func(c *ClientNodeOptions) *ClientNodeOptions

func NewClientNodeOptions() *ClientNodeOptions {
	return &ClientNodeOptions{
		options: []grpc.DialOption{},
	}
}

type ClientRetryOptionsInput struct {
	maxAttempts     uint
	backOff         time.Duration
	jitter          float64
	perRetryTimeout time.Duration
}

func WithClientRetryOptions(input ClientRetryOptionsInput) ClientNodeOption {
	return func(c *ClientNodeOptions) *ClientNodeOptions {
		opts := []grpc_retry.CallOption{
			grpc_retry.WithPerRetryTimeout(input.perRetryTimeout),
			grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(input.backOff, input.jitter)),
			grpc_retry.WithMax(input.maxAttempts),
		}

		c.options = append(c.options, grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))

		return c
	}
}

// WithDialOption sets the dial options for gRPC connections in the client
func WithInsecure() ClientNodeOption {
	return func(c *ClientNodeOptions) *ClientNodeOptions {
		c.options = append(c.options, grpc.WithInsecure())

		return c
	}
}

// WithDialOption sets the dial options for gRPC connections in the client
func WithDialOptions(option ...grpc.DialOption) ClientNodeOption {
	return func(c *ClientNodeOptions) *ClientNodeOptions {
		c.options = append(c.options, option...)

		return c
	}
}

func newClientNode(node registry.Node, currrentId string, optionFuncs ...ClientNodeOption) (*clientNode, error) {
	clientOptions := NewClientNodeOptions()
	for _, addOption := range optionFuncs {
		clientOptions = addOption(clientOptions)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Address, node.Port), clientOptions.options...)
	if err != nil {
		return nil, &ClientNodeDialError{
			Method:   "newClientNode",
			Err:      err,
			Port:     node.Port,
			Hostname: node.Address,
		}
	}

	n := clientNode{
		Node:       node,
		connection: conn,
		currentId:  currrentId,
	}

	return &n, nil
}

func (c *clientNode) sendMessage(ctx context.Context, m messaging.Message) error {
	var err error
	switch m.Type {
	case messaging.PubMessage:
		err = c.sendPublishMessage(ctx, m)
	case messaging.ReqMessage:
		err = c.sendRequestMessage(ctx, m)
	case messaging.RespMessage:
		err = c.sendResponseMessage(ctx, m)
	}

	if err != nil {
		return &ClientNodeSendError{
			Err:    err,
			Method: "sendMessage",
		}
	}

	return nil
}

func (c *clientNode) sendPublishMessage(ctx context.Context, m messaging.Message) error {
	if m.Type != messaging.PubMessage {
		return &ClientNodeMessageTypeError{
			Type:   messaging.PubMessage.String(),
			Method: "sendRequestMessage",
		}
	}

	_, err := proto.NewMessageClient(c.connection).PublishMessage(ctx, &proto.PublishMessageRequest{
		Message: &proto.PublishMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return &ClientNodeSendError{
			Err:    err,
			Method: "sendPublishMessage",
		}
	}

	return nil
}

func (c *clientNode) sendRequestMessage(ctx context.Context, m messaging.Message) error {
	if m.Type != messaging.ReqMessage {
		return &ClientNodeMessageTypeError{
			Type:   messaging.ReqMessage.String(),
			Method: "sendRequestMessage",
		}
	}

	_, err := proto.NewMessageClient(c.connection).RequestMessage(ctx, &proto.RequestMessageRequest{
		Message: &proto.RequestMessage{
			Id:      m.Id,
			NodeId:  c.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return &ClientNodeSendError{
			Err:    err,
			Method: "sendRequestMessage",
		}
	}

	return nil
}

func (c *clientNode) sendResponseMessage(ctx context.Context, m messaging.Message) error {
	if m.Type != messaging.RespMessage {
		return &ClientNodeMessageTypeError{
			Type:   messaging.RespMessage.String(),
			Method: "sendResponseMessage",
		}
	}

	_, err := proto.NewMessageClient(c.connection).ResponseMessage(ctx, &proto.ResponseMessageRequest{
		Message: &proto.ResponseMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return &ClientNodeSendError{
			Err:    err,
			Method: "sendResponseMessage",
		}
	}

	return nil
}

func (c *clientNode) Receiver() registry.Node {
	return c.Node
}
