package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type ClientOption func(m *MessageClient)

func WithDialOption(option ...grpc.DialOption) ClientOption {
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

func WithPublishSender(s Sender) ClientOption {
	return func(m *MessageClient) {
		m.publishSender = s
	}
}

func WithRequestSender(s Sender) ClientOption {
	return func(m *MessageClient) {
		m.requestSender = s
	}
}

func WithResponseSender(s Sender) ClientOption {
	return func(m *MessageClient) {
		m.responseSender = s
	}
}

type MessageClient struct {
	responseMap        *responseMap
	subscriberMap      *subscriberMap
	infoChannel        chan node.ResponseInfo
	pushChannel        chan message.Message
	nodeChannel        chan map[string]node.Node
	failedConnChannel  chan node.Node
	publishSender      Sender
	requestSender      Sender
	responseSender     Sender
	failedWaitInterval time.Duration
	maxFailedAttempts  int
	gRPCContext        context.Context
	id                 string
	dialOptions        []grpc.DialOption
}

func NewMessageClient(id string, info chan node.ResponseInfo, nchan chan map[string]node.Node, fcchan chan node.Node, options ...ClientOption) *MessageClient {
	c := &MessageClient{
		responseMap:        newResponseMap(),
		subscriberMap:      newSubscriberMap(),
		infoChannel:        info,
		pushChannel:        make(chan message.Message),
		nodeChannel:        nchan,
		publishSender:      &PublishClient{},
		requestSender:      &RequestClient{NodeId: id},
		responseSender:     &ResponseClient{},
		failedConnChannel:  fcchan,
		gRPCContext:        context.Background(),
		dialOptions:        []grpc.DialOption{},
		failedWaitInterval: time.Second * 3,
		maxFailedAttempts:  5,
	}

	for _, option := range options {
		option(c)
	}

	go c.listenForResponseInfo()
	go c.listenForSubscribers()
	go c.listenForOutgoingMessages()

	return c
}

func (c *MessageClient) PushChannel() chan message.Message {
	return c.pushChannel
}

func (c *MessageClient) listenForResponseInfo() {
	for response := range c.infoChannel {
		c.responseMap.PushResponse(response)
	}
}

func (c *MessageClient) listenForSubscribers() {
	for subs := range c.nodeChannel {
		count := 0
		for _, sub := range subs {
			count++
			c.subscriberMap.AddSubscriber(sub)
		}
	}
}

func (c *MessageClient) listenForOutgoingMessages() {
	for m := range c.pushChannel {
		go func(m message.Message) {
			switch m.Type {
			case message.PubMessage:
				nodes, err := c.subscriberMap.SubjectSubscribers(m.Subject)
				if err != nil {
					log.Printf("could not get subject subscribers: %s", err)
				}

				for _, n := range nodes {
					go c.attemptMessage(m, n, c.publishSender)
				}
			case message.ReqMessage:
				nodes, err := c.subscriberMap.SubjectSubscribers(m.Subject)
				if err != nil {
					log.Printf("could not get subject subscribers: %s", err)
				}

				for _, n := range nodes {
					go c.attemptMessage(m, n, c.requestSender)
				}
			case message.RespMessage:
				id, err := c.responseMap.PopResponse(m.Id)
				if err != nil {
					log.Printf("could not get node id for message id %s: %s", m.Id, err)
				}

				n, err := c.subscriberMap.Subscriber(id)
				if err != nil {
					log.Printf("could node with id %s: %s", id, err)
				}

				go c.attemptMessage(m, n, c.responseSender)
			}
		}(m)
	}
}

func (c *MessageClient) attemptMessage(m message.Message, receiver *node.Node, sender Sender) {
	attempt := 0
	for {
		client, conn, err := createGRPCClient(receiver.Address, receiver.Port, c.dialOptions...)
		if err == nil {
			err = sender.Send(c.gRPCContext, m, client)
			if err == nil {
				conn.Close()
				break
			}
		}

		conn.Close()
		if attempt >= c.maxFailedAttempts {
			log.Printf("failed creating connection to %s:%s: %s\n", receiver.Address, receiver.Port, err)
			log.Printf("will now be blacklisting %s\n", receiver.Id)

			c.failedConnChannel <- *receiver
			break
		}

		time.Sleep(c.failedWaitInterval)
		attempt++
		continue
	}
}

func createGRPCClient(address string, port string, options ...grpc.DialOption) (proto.MessageServerClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), options...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial %s: %v", fmt.Sprintf("%s:%s", address, port), err)
	}

	client := proto.NewMessageServerClient(conn)

	return client, conn, nil
}
