package messaging

import (
	"context"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/platform-edn/courier/messaging/proto"
	"google.golang.org/grpc"
)

type clientNode struct {
	Node
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

func newClientNode(node Node, currrentId string, optionFuncs ...ClientNodeOption) (*clientNode, error) {
	clientOptions := NewClientNodeOptions()
	for _, addOption := range optionFuncs {
		clientOptions = addOption(clientOptions)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.address, node.port), clientOptions.options...)
	if err != nil {
		return nil, &ClientNodeDialError{
			Method:   "newClientNode",
			Err:      err,
			Port:     node.port,
			Hostname: node.address,
		}
	}

	n := clientNode{
		Node:       node,
		connection: conn,
		currentId:  currrentId,
	}

	return &n, nil
}

func (c *clientNode) sendMessage(ctx context.Context, m Message) error {
	var err error
	switch m.Type {
	case PubMessage:
		err = c.sendPublishMessage(ctx, m)
	case ReqMessage:
		err = c.sendRequestMessage(ctx, m)
	case RespMessage:
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

func (c *clientNode) sendPublishMessage(ctx context.Context, m Message) error {
	if m.Type != PubMessage {
		return &ClientNodeMessageTypeError{
			Type:   PubMessage,
			Method: "sendRequestMessage",
		}
	}

	_, err := proto.NewMessageServerClient(c.connection).PublishMessage(ctx, &proto.PublishMessageRequest{
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

func (c *clientNode) sendRequestMessage(ctx context.Context, m Message) error {
	if m.Type != ReqMessage {
		return &ClientNodeMessageTypeError{
			Type:   ReqMessage,
			Method: "sendRequestMessage",
		}
	}

	_, err := proto.NewMessageServerClient(c.connection).RequestMessage(ctx, &proto.RequestMessageRequest{
		Message: &proto.RequestMessage{
			Id:      m.Id,
			NodeId:  c.id,
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

func (c *clientNode) sendResponseMessage(ctx context.Context, m Message) error {
	if m.Type != RespMessage {
		return &ClientNodeMessageTypeError{
			Type:   RespMessage,
			Method: "sendResponseMessage",
		}
	}

	_, err := proto.NewMessageServerClient(c.connection).ResponseMessage(ctx, &proto.ResponseMessageRequest{
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

func (c *clientNode) Receiver() Node {
	return c.Node
}
