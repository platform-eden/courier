package server

import (
	"context"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
)

type Server interface {
	PushChannel() chan message.Message
	ResponseChannel() chan node.ResponseInfo
}

type MessageServer struct {
	pushChannel     chan message.Message
	responseChannel chan node.ResponseInfo
	proto.UnimplementedMessageServerServer
}

func NewMessageServer() *MessageServer {
	m := MessageServer{
		pushChannel:     make(chan message.Message),
		responseChannel: make(chan node.ResponseInfo),
	}

	return &m
}

func (m *MessageServer) PushChannel() chan message.Message {
	return m.pushChannel
}

func (m *MessageServer) ResponseChannel() chan node.ResponseInfo {
	return m.responseChannel
}

func (m *MessageServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	pub := message.NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	m.pushChannel <- pub

	response := proto.PublishMessageResponse{}

	return &response, nil
}

func (m *MessageServer) RequestMessage(ctx context.Context, request *proto.RequestMessageRequest) (*proto.RequestMessageResponse, error) {
	req := message.NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := node.ResponseInfo{
		MessageId: request.Message.Id,
		NodeId:    request.Message.NodeId,
	}

	m.pushChannel <- req
	m.responseChannel <- info

	response := proto.RequestMessageResponse{}

	return &response, nil
}

func (m *MessageServer) ResponseMessage(ctx context.Context, request *proto.ResponseMessageRequest) (*proto.ResponseMessageResponse, error) {
	resp := message.NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	m.pushChannel <- resp

	response := proto.ResponseMessageResponse{}

	return &response, nil
}
