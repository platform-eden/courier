package client

import (
	"context"
	"fmt"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/proto"
)

type Sender interface {
	Send(context.Context, message.Message, proto.MessageServerClient) error
}

type PublishClient struct{}

func (p *PublishClient) Send(ctx context.Context, m message.Message, client proto.MessageServerClient) error {
	if m.Type != message.PubMessage {
		return fmt.Errorf("message type must be of type PublishMessage")
	}

	_, err := client.PublishMessage(ctx, &proto.PublishMessageRequest{
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

type RequestClient struct {
	NodeId string
}

func (r *RequestClient) Send(ctx context.Context, m message.Message, client proto.MessageServerClient) error {
	if m.Type != message.ReqMessage {
		return fmt.Errorf("message type must be of type RequestMessage")
	}

	_, err := client.RequestMessage(ctx, &proto.RequestMessageRequest{
		Message: &proto.RequestMessage{
			Id:      m.Id,
			NodeId:  r.NodeId,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

type ResponseClient struct{}

func (r *ResponseClient) Send(ctx context.Context, m message.Message, client proto.MessageServerClient) error {
	if m.Type != message.RespMessage {
		return fmt.Errorf("message type must be of type ResponseMessage")
	}
	_, err := client.ResponseMessage(ctx, &proto.ResponseMessageRequest{
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
