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
	"google.golang.org/grpc/credentials/insecure"
)

type clientNode struct {
	registry.Node
	connection *grpc.ClientConn
	currentId  string
}

type ClientNodeOptions struct {
	options []grpc.DialOption
}

type ClientNodeOption func(c *ClientNodeOptions) *ClientNodeOptions

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
		c.options = append(c.options, grpc.WithTransportCredentials(insecure.NewCredentials()))

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
	clientOptions := &ClientNodeOptions{}
	for _, addOption := range optionFuncs {
		clientOptions = addOption(clientOptions)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Address, node.Port), clientOptions.options...)
	if err != nil {
		return nil, fmt.Errorf("NewClientNode: %w", err)
	}

	n := &clientNode{
		Node:       node,
		connection: conn,
		currentId:  currrentId,
	}

	return n, nil
}

func (c *clientNode) SendPublishMessage(ctx context.Context, m messaging.Message, failedNodes chan registry.Node) {
	_, err := proto.NewMessageClient(c.connection).PublishMessage(ctx, &proto.PublishMessageRequest{
		Message: &proto.PublishMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		if err != nil {
			failedNodes <- c.Node
		}
	}
}

func (c *clientNode) SendRequestMessage(ctx context.Context, m messaging.Message, failedNodes chan registry.Node) {
	_, err := proto.NewMessageClient(c.connection).RequestMessage(ctx, &proto.RequestMessageRequest{
		Message: &proto.RequestMessage{
			Id:      m.Id,
			NodeId:  c.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		if err != nil {
			failedNodes <- c.Node
		}
	}
}

func (c *clientNode) SendResponseMessage(ctx context.Context, m messaging.Message, failedNodes chan registry.Node) {
	_, err := proto.NewMessageClient(c.connection).ResponseMessage(ctx, &proto.ResponseMessageRequest{
		Message: &proto.ResponseMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		if err != nil {
			failedNodes <- c.Node
		}
	}
}
