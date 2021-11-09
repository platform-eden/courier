package courier

import (
	"context"
	"fmt"
	"sync"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
)

type MessageServer struct {
	responseChannel chan node.ResponseInfo
	channelMap      channelMapper
	proto.UnimplementedMessageServerServer
}

func NewMessageServer(rchan chan node.ResponseInfo, chanMap channelMapper) *MessageServer {
	m := MessageServer{
		responseChannel: rchan,
		channelMap:      chanMap,
	}

	return &m
}

func (m *MessageServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	pub := message.NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := m.channelMap.Subscriptions(pub.Subject)
	if err != nil {
		return nil, fmt.Errorf("could not get subscriptions for %s: %s", pub.Subject, err)
	}

	cchan := generateMessageChannels(channels)
	fanForwardMessages(cchan, pub, forwardMessage)

	response := proto.PublishMessageResponse{}

	return &response, nil
}

func (m *MessageServer) RequestMessage(ctx context.Context, request *proto.RequestMessageRequest) (*proto.RequestMessageResponse, error) {
	req := message.NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := node.ResponseInfo{
		MessageId: request.Message.Id,
		NodeId:    request.Message.NodeId,
	}

	m.responseChannel <- info

	channels, err := m.channelMap.Subscriptions(req.Subject)
	if err != nil {
		return nil, fmt.Errorf("could not get subscriptions for %s: %s", req.Subject, err)
	}

	cchan := generateMessageChannels(channels)
	fanForwardMessages(cchan, req, forwardMessage)

	response := proto.RequestMessageResponse{}

	return &response, nil
}

func (m *MessageServer) ResponseMessage(ctx context.Context, request *proto.ResponseMessageRequest) (*proto.ResponseMessageResponse, error) {
	resp := message.NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := m.channelMap.Subscriptions(resp.Subject)
	if err != nil {
		return nil, fmt.Errorf("could not get subscriptions for %s: %s", resp.Subject, err)
	}

	cchan := generateMessageChannels(channels)
	fanForwardMessages(cchan, resp, forwardMessage)

	response := proto.ResponseMessageResponse{}

	return &response, nil
}

func generateMessageChannels(channels []chan message.Message) <-chan chan message.Message {
	out := make(chan chan message.Message)

	go func() {
		for _, channel := range channels {
			out <- channel
		}

		close(out)
	}()

	return out
}

type forwardFunc func(mchan chan message.Message, m message.Message, wg *sync.WaitGroup)

func fanForwardMessages(cchan <-chan chan message.Message, m message.Message, forward forwardFunc) {
	wg := &sync.WaitGroup{}

	for channel := range cchan {
		wg.Add(1)
		go forward(channel, m, wg)
	}

	wg.Wait()
}

func forwardMessage(mchan chan message.Message, m message.Message, wg *sync.WaitGroup) {
	defer wg.Done()

	mchan <- m
}
