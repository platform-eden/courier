package server

import (
	"context"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
)

type MessageServer struct {
	pushChannel chan message.Message
	infoChannel chan node.ResponseInfo
	proto.UnimplementedMessageServerServer
}

func NewMessageServer(push chan message.Message, info chan node.ResponseInfo) *MessageServer {
	m := MessageServer{
		pushChannel: push,
		infoChannel: info,
	}

	return &m
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
		Id:      request.Message.Id,
		Address: request.Message.ReturnAddress,
		Port:    request.Message.ReturnPort,
	}

	m.pushChannel <- req
	m.infoChannel <- info

	response := proto.RequestMessageResponse{}

	return &response, nil
}

func (m *MessageServer) ResponseMessage(ctx context.Context, request *proto.ResponseMessageRequest) (*proto.ResponseMessageResponse, error) {
	resp := message.NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	m.pushChannel <- resp

	response := proto.ResponseMessageResponse{}

	return &response, nil
}
