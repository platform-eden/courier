package courier

import (
	"context"
	"fmt"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type clientNode struct {
	Node
	connection *grpc.ClientConn
	currentId  string
}

func newClientNode(node Node, currrentId string, options ...grpc.DialOption) (*clientNode, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Address, node.Port), options...)
	if err != nil {
		return nil, fmt.Errorf("could not create connection at %s: %s", fmt.Sprintf("%s:%s", node.Address, node.Port), err)
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
		return fmt.Errorf("error semding message: %s", err)
	}

	return nil
}

func (c *clientNode) sendPublishMessage(ctx context.Context, m Message) error {
	if m.Type != PubMessage {
		return fmt.Errorf("message type must be of type PublishMessage")
	}

	_, err := proto.NewMessageServerClient(c.connection).PublishMessage(ctx, &proto.PublishMessageRequest{
		Message: &proto.PublishMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

func (c *clientNode) sendRequestMessage(ctx context.Context, m Message) error {
	if m.Type != ReqMessage {
		return fmt.Errorf("message type must be of type RequestMessage")
	}

	_, err := proto.NewMessageServerClient(c.connection).RequestMessage(ctx, &proto.RequestMessageRequest{
		Message: &proto.RequestMessage{
			Id:      m.Id,
			NodeId:  c.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

func (c *clientNode) sendResponseMessage(ctx context.Context, m Message) error {
	if m.Type != RespMessage {
		return fmt.Errorf("message type must be of type ResponseMessage")
	}

	_, err := proto.NewMessageServerClient(c.connection).ResponseMessage(ctx, &proto.ResponseMessageRequest{
		Message: &proto.ResponseMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

func (c *clientNode) Receiver() Node {
	return c.Node
}
