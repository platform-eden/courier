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

type ClientNodeDialError struct {
	Hostname string
	Port     string
	Err      error
}

func (err *ClientNodeDialError) Error() string {
	return fmt.Sprintf("newClientNode: could not create connection at %s:%s: %s", err.Hostname, err.Port, err.Err)
}

type ClientNodeSendError struct {
	Method string
	Err    error
}

func (err *ClientNodeSendError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type ClientNodeMessageTypeError struct {
	Method string
	Type   messageType
}

func (err *ClientNodeMessageTypeError) Error() string {
	return fmt.Sprintf("%s: message must be of type %s", err.Method, err.Type)
}

func newClientNode(node Node, currrentId string, options ...grpc.DialOption) (*clientNode, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.address, node.port), options...)
	if err != nil {
		return nil, &ClientNodeDialError{
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
