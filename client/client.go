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

func WithAddress(address string) ClientOption {
	return func(m *MessageClient) {
		m.Address = address
	}
}

func WithPort(port string) ClientOption {
	return func(m *MessageClient) {
		m.Port = port
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

type MessageClient struct {
	responseMap        *responseMap
	subscriberMap      *subscriberMap
	infoChannel        chan node.ResponseInfo
	pushChannel        chan message.Message
	nodeChannel        chan map[string]node.Node
	failedConnChannel  chan node.Node
	failedWaitInterval time.Duration
	maxFailedAttempts  int
	gRPCContext        context.Context
	Address            string
	Port               string
	dialOptions        []grpc.DialOption
}

func NewMessageClient(info chan node.ResponseInfo, node chan map[string]node.Node, fcchan chan node.Node, options ...ClientOption) *MessageClient {
	c := &MessageClient{
		responseMap:        newResponseMap(),
		subscriberMap:      newSubscriberMap(),
		infoChannel:        info,
		pushChannel:        make(chan message.Message),
		nodeChannel:        node,
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
			var err error
			switch m.Type {
			case message.PubMessage:
				err = sendPubMessage(&m, c.subscriberMap, c.gRPCContext, c.dialOptions)
				if err != nil {
					log.Printf("failed sending PubMessage: %s", err)
				}
			case message.RespMessage:
				err = sendRespMessage(&m, c.responseMap, c.gRPCContext, c.dialOptions)
				if err != nil {
					log.Printf("failed sending RespMessage: %s", err)
				}
			case message.ReqMessage:
				err = sendReqMessage(&m, c.subscriberMap, c.Address, c.Port, c.gRPCContext, c.dialOptions)
				if err != nil {
					log.Printf("failed sending ReqMessage: %s", err)
				}
			}
		}(m)
	}
}

func sendPubMessage(m *message.Message, subs *subscriberMap, ctx context.Context, options []grpc.DialOption) error {
	nodes, err := subs.SubjectSubscribers(m.Subject)
	if err != nil {
		return fmt.Errorf("could not get subject subscribers: %s", err)
	}

	for _, n := range nodes {
		go func(n *node.Node) {
			client, conn, err := createGRPCClient(ctx, n.Address, n.Port, options...)
			if err != nil {
				log.Printf("failed creating grpc client for %s:%s : %s", n.Address, n.Port, err)
				return
			}
			defer conn.Close()

			p := proto.PublishMessage{
				Id:      m.Id,
				Subject: m.Subject,
				Content: m.Content,
			}

			_, err = client.PublishMessage(ctx, &proto.PublishMessageRequest{Message: &p})
			if err != nil {
				log.Printf("failed sending publish message to %s: %s", n.Address, err)
			}
		}(n)
	}

	return nil
}

func sendReqMessage(m *message.Message, subs *subscriberMap, address string, port string, ctx context.Context, options []grpc.DialOption) error {
	nodes, err := subs.SubjectSubscribers(m.Subject)
	if err != nil {
		return fmt.Errorf("could not get subject subscribers: %s", err)
	}

	for _, n := range nodes {
		go func(n *node.Node) {
			client, conn, err := createGRPCClient(ctx, n.Address, n.Port, options...)
			if err != nil {
				log.Printf("failed creating grpc client for %s:%s : %s", n.Address, n.Port, err)
				return
			}
			defer conn.Close()

			p := proto.RequestMessage{
				Id:            m.Id,
				ReturnAddress: address,
				ReturnPort:    port,
				Subject:       m.Subject,
				Content:       m.Content,
			}

			_, err = client.RequestMessage(ctx, &proto.RequestMessageRequest{Message: &p})
			if err != nil {
				log.Printf("failed sending publish message to %s: %s", n.Address, err)
			}
		}(n)
	}

	return nil
}

func sendRespMessage(m *message.Message, resps *responseMap, ctx context.Context, options []grpc.DialOption) error {
	info, err := resps.PopResponse(m.Id)
	if err != nil {
		return fmt.Errorf("couldn't get pop response: %s", err)
	}

	client, conn, err := createGRPCClient(ctx, info.Address, info.Port, options...)
	if err != nil {
		return fmt.Errorf("failed creating grpc client for %s:%s : %s", info.Address, info.Port, err)
	}
	defer conn.Close()

	resp := &proto.ResponseMessage{
		Id:      m.Id,
		Subject: m.Subject,
		Content: m.Content,
	}

	_, err = client.ResponseMessage(ctx, &proto.ResponseMessageRequest{Message: resp})
	if err != nil {
		return fmt.Errorf("failed sending response to %s:%s : %s", info.Address, info.Port, err)
	}
	return nil
}

func createGRPCClient(ctx context.Context, address string, port string, options ...grpc.DialOption) (proto.MessageServerClient, *grpc.ClientConn, error) {

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%s", address, port), options...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial %s: %v", fmt.Sprintf("%s:%s", address, port), err)
	}

	client := proto.NewMessageServerClient(conn)

	return client, conn, nil
}
