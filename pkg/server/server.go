package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/platform-edn/courier/pkg/messaging"
)

type ChannelMapper interface {
	SubscribeToSubject(string) <-chan messaging.Message
	Subscriptions(string) ([]chan messaging.Message, error)
	CloseSubscriberChannels()
	GenerateMessageChannels(string) (<-chan chan messaging.Message, error)
}

type MessagingServer struct {
	responseChannel chan messaging.ResponseInfo
	ChannelMapper
	messaging.UnimplementedMessageServer
}

func NewMessagingServer() *MessagingServer {
	s := &MessagingServer{
		responseChannel: make(chan messaging.ResponseInfo),
		ChannelMapper:   NewChannelMap(),
	}

	return s
}

func (s *MessagingServer) ResponseChannel() <-chan messaging.ResponseInfo {
	return s.responseChannel
}

func (s *MessagingServer) PublishMessage(ctx context.Context, request *messaging.PublishMessageRequest) (*messaging.PublishMessageResponse, error) {
	pub := messaging.NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	err := s.FanForwardMessages(ctx, pub.Subject, pub)
	if err != nil {
		return nil, fmt.Errorf("PublishMessage: %w", err)
	}

	response := messaging.PublishMessageResponse{}

	return &response, nil
}

func (s *MessagingServer) RequestMessage(ctx context.Context, request *messaging.RequestMessageRequest) (*messaging.RequestMessageResponse, error) {
	req := messaging.NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	s.responseChannel <- messaging.ResponseInfo{
		MessageId: request.Message.Id,
		NodeId:    request.Message.NodeId,
	}

	err := s.FanForwardMessages(ctx, req.Subject, req)
	if err != nil {
		return nil, fmt.Errorf("RequestMessage: %w", err)
	}

	response := messaging.RequestMessageResponse{}

	return &response, nil
}

func (s *MessagingServer) ResponseMessage(ctx context.Context, request *messaging.ResponseMessageRequest) (*messaging.ResponseMessageResponse, error) {
	resp := messaging.NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	err := s.FanForwardMessages(ctx, resp.Subject, resp)
	if err != nil {
		return nil, fmt.Errorf("ResponseMessage: %w", err)
	}

	response := messaging.ResponseMessageResponse{}

	return &response, nil
}

func (s *MessagingServer) FanForwardMessages(ctx context.Context, subject string, msg messaging.Message) error {
	wg := &sync.WaitGroup{}

	channels, err := s.GenerateMessageChannels(subject)
	if err != nil {
		return fmt.Errorf("FanForwardMessage: %w", err)
	}

	for channel := range channels {
		wg.Add(1)
		go func(channel chan messaging.Message) {
			defer wg.Done()

			channel <- msg
		}(channel)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("FanForwardMessages: %w", &ForwardMessagingError{
			Message: msg.Id,
			Subject: subject,
		})
	}

	return nil
}
