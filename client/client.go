package client

import (
	"context"
	"fmt"
	"time"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"google.golang.org/grpc"
)

type ClientOption func(m *MessageClient)

func WithId(id string) ClientOption {
	return func(m *MessageClient) {
		m.Id = id
		m.Senders[RequestSender] = &RequestClient{NodeId: id}
	}
}

func WithInfoChannel(channel chan node.ResponseInfo) ClientOption {
	return func(m *MessageClient) {
		m.infoChannel = channel
	}
}

func WithFailedConnectionChannel(channel chan node.Node) ClientOption {
	return func(m *MessageClient) {
		m.failedConnChannel = channel
	}
}

func WithNodeChannel(channel chan map[string]node.Node) ClientOption {
	return func(m *MessageClient) {
		m.nodeChannel = channel
	}
}

func WithDialOptions(option ...grpc.DialOption) ClientOption {
	return func(m *MessageClient) {
		m.dialOptions = append(m.dialOptions, option...)
	}
}

func WithContext(ctx context.Context) ClientOption {
	return func(m *MessageClient) {
		m.gRPCContext = ctx
	}
}

func WithFailedWaitInterval(interval time.Duration) ClientOption {
	return func(m *MessageClient) {
		m.failedWaitInterval = interval
	}
}

func WithMaxFailedAttempts(attempts int) ClientOption {
	return func(m *MessageClient) {
		m.maxFailedAttempts = attempts
	}
}

func WithSender(t senderType, s Sender) ClientOption {
	return func(m *MessageClient) {
		m.Senders[t] = s
	}
}

type MessageClient struct {
	Id                 string
	responseMap        *responseMap
	subscriberMap      *subscriberMap
	infoChannel        chan node.ResponseInfo
	messageChannel     chan message.Message
	nodeChannel        chan map[string]node.Node
	failedConnChannel  chan node.Node
	Senders            map[senderType]Sender
	failedWaitInterval time.Duration
	maxFailedAttempts  int
	gRPCContext        context.Context
	dialOptions        []grpc.DialOption
}

func NewMessageClient(options ...ClientOption) (*MessageClient, error) {
	c := &MessageClient{
		Id:             "",
		responseMap:    newResponseMap(),
		subscriberMap:  newSubscriberMap(),
		infoChannel:    nil,
		messageChannel: make(chan message.Message),
		nodeChannel:    nil,
		Senders: map[senderType]Sender{
			PublishSender:  &PublishClient{},
			RequestSender:  &RequestClient{},
			ResponseSender: &ResponseClient{},
		},
		failedConnChannel:  nil,
		gRPCContext:        context.Background(),
		dialOptions:        []grpc.DialOption{},
		failedWaitInterval: time.Second * 3,
		maxFailedAttempts:  5,
	}

	for _, option := range options {
		option(c)

	}

	if c.Id == "" {
		return nil, fmt.Errorf("id cannot be empty string")
	} else if c.infoChannel == nil {
		return nil, fmt.Errorf("info channel cannot be nil")
	} else if c.nodeChannel == nil {
		return nil, fmt.Errorf("node channel cannot be nil")
	}

	return c, nil
}

func (c *MessageClient) InfoChannel() chan node.ResponseInfo {
	return c.infoChannel
}

func (c *MessageClient) NodeChannel() chan map[string]node.Node {
	return c.nodeChannel
}

func (c *MessageClient) MessageChannel() chan message.Message {
	return c.messageChannel
}

func (c *MessageClient) FailedConnectionChannel() chan node.Node {
	return c.failedConnChannel
}

func (c *MessageClient) DialOptions() []grpc.DialOption {
	return c.dialOptions
}

func (c *MessageClient) GRPCContext() context.Context {
	return c.gRPCContext
}

func (c *MessageClient) MaxFailedAttempts() int {
	return c.maxFailedAttempts
}

func (c *MessageClient) FailedWaitInterval() time.Duration {
	return c.failedWaitInterval
}

func (c *MessageClient) Sender(sender senderType) Sender {
	return c.Senders[sender]
}

func (c *MessageClient) PushResponse(response node.ResponseInfo) {
	c.responseMap.PushResponse(response)
}

func (c *MessageClient) UpdateNodes(nodes map[string]node.Node) {
	for _, n := range nodes {
		c.subscriberMap.AddSubscriber(n)
	}
}

func (c *MessageClient) SubjectSubscribers(subject string) ([]*node.Node, error) {
	subscribers, err := c.subscriberMap.SubjectSubscribers(subject)
	if err != nil {
		return nil, err
	}

	return subscribers, nil
}

func (c *MessageClient) Response(id string) (string, error) {
	nodeId, err := c.responseMap.PopResponse(id)
	if err != nil {
		return "", err
	}

	return nodeId, nil
}

func (c *MessageClient) Subscriber(id string) (*node.Node, error) {
	subscriber, err := c.subscriberMap.Subscriber(id)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (c *MessageClient) RemoveSubscriber(subscriber *node.Node) error {
	err := c.subscriberMap.RemoveSubscriber(subscriber)
	if err != nil {
		return err
	}

	return nil
}
