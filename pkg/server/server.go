package server

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/messaging/proto"
	"google.golang.org/grpc"
)

type channelMapper interface {
	AddSubscriber(string) <-chan messaging.Message
	Subscriptions(string) ([]chan messaging.Message, error)
	Close()
}

type messagingServer struct {
	responseChannel chan messaging.ResponseInfo
	channelMapper
	proto.UnimplementedMessageServer
}

func NewMessagingServer() *messagingServer {
	s := &messagingServer{
		responseChannel: make(chan messaging.ResponseInfo),
		channelMapper:   newChannelMap(),
	}

	// if options.startServer {
	// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	// 	if err != nil {
	// 		return nil, &messagingServerStartError{
	// 			Method: "Start",
	// 			Err:    err,
	// 		}
	// 	}

	// 	server := grpc.NewServer()
	// 	proto.RegisterMessageServerServer(server, s)

	// 	go startMessagingServer(ctx, wg, server, lis, s.port)
	// }

	return s
}

func (s *messagingServer) SubscribeToSubject(subject string) <-chan messaging.Message {
	channel := s.AddSubscriber(subject)

	return channel
}

func (s *messagingServer) ResponseChannel() <-chan messaging.ResponseInfo {
	return s.responseChannel
}

func (s *messagingServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	pub := messaging.NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := s.Subscriptions(pub.Subject)
	if err != nil {
		return nil, &ChannelSubscriptionError{
			Method: "PublishMessage",
			Err:    err,
		}
	}

	cchan := generateMessageChannels(channels)
	fanForwardMessages(cchan, pub, forwardMessage)

	response := proto.PublishMessageResponse{}

	return &response, nil
}

func (s *messagingServer) RequestMessage(ctx context.Context, request *proto.RequestMessageRequest) (*proto.RequestMessageResponse, error) {
	req := messaging.NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := messaging.ResponseInfo{
		MessageId: request.Message.Id,
		NodeId:    request.Message.NodeId,
	}

	s.responseChannel <- info

	channels, err := s.Subscriptions(req.Subject)
	if err != nil {
		return nil, &ChannelSubscriptionError{
			Method: "RequestMessage",
			Err:    err,
		}
	}

	cchan := generateMessageChannels(channels)
	fanForwardMessages(cchan, req, forwardMessage)

	response := proto.RequestMessageResponse{}

	return &response, nil
}

func (s *messagingServer) ResponseMessage(ctx context.Context, request *proto.ResponseMessageRequest) (*proto.ResponseMessageResponse, error) {
	resp := messaging.NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := s.Subscriptions(resp.Subject)
	if err != nil {
		return nil, &ChannelSubscriptionError{
			Method: "ResponseMessage",
			Err:    err,
		}
	}

	cchan := generateMessageChannels(channels)
	fanForwardMessages(cchan, resp, forwardMessage)

	response := proto.ResponseMessageResponse{}

	return &response, nil
}

func generateMessageChannels(channels []chan messaging.Message) <-chan chan messaging.Message {
	out := make(chan chan messaging.Message)

	go func() {
		for _, channel := range channels {
			out <- channel
		}

		close(out)
	}()

	return out
}

type forwardFunc func(mchan chan messaging.Message, m messaging.Message, wg *sync.WaitGroup)

func fanForwardMessages(cchan <-chan chan messaging.Message, m messaging.Message, forward forwardFunc) {
	wg := &sync.WaitGroup{}

	for channel := range cchan {
		wg.Add(1)
		go forward(channel, m, wg)
	}

	wg.Wait()
}

func forwardMessage(mchan chan messaging.Message, m messaging.Message, wg *sync.WaitGroup) {
	defer wg.Done()

	mchan <- m
}

// localIP returns the ip address this node is currently using
func localIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// startmessagingServer starts the message server on a given port
func startMessagingServer(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, lis net.Listener, port string) {
	errchan := make(chan error, 1)
	done := make(chan struct{})
	defer close(errchan)
	defer wg.Done()

	go func() {
		err := server.Serve(lis)
		if err != nil {
			errchan <- err
		}
		defer close(done)
	}()

	select {
	case <-ctx.Done():
		server.GracefulStop()
		<-done
	case err := <-errchan:
		log.Print(err.Error())
		os.Exit(1)
	}
}
