package client

import (
	"context"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientNode struct {
	registry.Node
	messaging.MessageClient
	currentId string
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

func NewClientNode(node registry.Node, currrentId string, optionFuncs ...ClientNodeOption) (*ClientNode, error) {
	clientOptions := &ClientNodeOptions{}
	for _, addOption := range optionFuncs {
		clientOptions = addOption(clientOptions)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Address, node.Port), clientOptions.options...)
	if err != nil {
		return nil, fmt.Errorf("NewClientNode: %w", err)
	}

	n := &ClientNode{
		Node:          node,
		MessageClient: messaging.NewMessageClient(conn),
		currentId:     currrentId,
	}

	return n, nil
}

func (client *ClientNode) AttemptMessage(ctx context.Context, msg messaging.Message) error {
	var err error
	switch msg.Type {
	case messaging.PubMessage:
		_, err = client.PublishMessage(ctx, &messaging.PublishMessageRequest{
			Message: &messaging.PublishMessage{
				Id:      msg.Id,
				Subject: msg.Subject,
				Content: msg.Content,
			},
		})
	case messaging.ReqMessage:
		_, err = client.RequestMessage(ctx, &messaging.RequestMessageRequest{
			Message: &messaging.RequestMessage{
				Id:      msg.Id,
				NodeId:  client.currentId,
				Subject: msg.Subject,
				Content: msg.Content,
			},
		})
	case messaging.RespMessage:
		_, err = client.ResponseMessage(ctx, &messaging.ResponseMessageRequest{
			Message: &messaging.ResponseMessage{
				Id:      msg.Id,
				Subject: msg.Subject,
				Content: msg.Content,
			},
		})
	}

	if err != nil {
		return fmt.Errorf("AttemptMessage: %w", err)
	}

	return nil
}

func (client *ClientNode) Subscriber() registry.Node {
	return client.Node
}

// func (c *ClientNode) SendPublishMessage(ctx context.Context, m messaging.Message) error {
// 	_, err := c.PublishMessage(ctx, &messaging.PublishMessageRequest{
// 		Message: &messaging.PublishMessage{
// 			Id:      m.Id,
// 			Subject: m.Subject,
// 			Content: m.Content,
// 		},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("SendPublishMessage: %w", err)
// 	}

// 	return nil
// }

// func (c *ClientNode) SendRequestMessage(ctx context.Context, m messaging.Message) error {
// 	_, err := c.RequestMessage(ctx, &messaging.RequestMessageRequest{
// 		Message: &messaging.RequestMessage{
// 			Id:      m.Id,
// 			NodeId:  c.Id,
// 			Subject: m.Subject,
// 			Content: m.Content,
// 		},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("SendRequestMessage: %w", err)
// 	}

// 	return nil
// }

// func (c *ClientNode) SendResponseMessage(ctx context.Context, m messaging.Message) error {
// 	_, err := c.ResponseMessage(ctx, &messaging.ResponseMessageRequest{
// 		Message: &messaging.ResponseMessage{
// 			Id:      m.Id,
// 			Subject: m.Subject,
// 			Content: m.Content,
// 		},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("SendResponseMessage: %w", err)
// 	}

// 	return nil
// }
