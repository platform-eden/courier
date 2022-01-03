package messaging

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/platform-edn/courier/messaging/proto"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	port            string
	responseChannel chan ResponseInfo
	channelMap      channelMapper
	proto.UnimplementedMessageServerServer
}

type ChannelSubscriptionError struct {
	Method string
	Err    error
}

func (err *ChannelSubscriptionError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type gRPCServerStartError struct {
	Method string
	Err    error
}

func (err *gRPCServerStartError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

func NewgRPCServer(port string, rchan chan ResponseInfo, chanMap channelMapper) *gRPCServer {
	m := gRPCServer{
		port:            port,
		responseChannel: rchan,
		channelMap:      chanMap,
	}

	return &m
}

func (s *gRPCServer) start(ctx context.Context, wg *sync.WaitGroup) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return &gRPCServerStartError{
			Method: "Start",
			Err:    err,
		}
	}

	server := grpc.NewServer()
	proto.RegisterMessageServerServer(server, s)

	go startCourierServer(ctx, wg, server, lis, s.port)

	return nil
}

func (s *gRPCServer) subscribe(subject string) <-chan Message {
	return s.channelMap.Add(subject)
}

func (s *gRPCServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	pub := NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := s.channelMap.Subscriptions(pub.Subject)
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

func (s *gRPCServer) RequestMessage(ctx context.Context, request *proto.RequestMessageRequest) (*proto.RequestMessageResponse, error) {
	req := NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := ResponseInfo{
		MessageId: request.Message.Id,
		NodeId:    request.Message.NodeId,
	}

	s.responseChannel <- info

	channels, err := s.channelMap.Subscriptions(req.Subject)
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

func (s *gRPCServer) ResponseMessage(ctx context.Context, request *proto.ResponseMessageRequest) (*proto.ResponseMessageResponse, error) {
	resp := NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := s.channelMap.Subscriptions(resp.Subject)
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

func generateMessageChannels(channels []chan Message) <-chan chan Message {
	out := make(chan chan Message)

	go func() {
		for _, channel := range channels {
			out <- channel
		}

		close(out)
	}()

	return out
}

type forwardFunc func(mchan chan Message, m Message, wg *sync.WaitGroup)

func fanForwardMessages(cchan <-chan chan Message, m Message, forward forwardFunc) {
	wg := &sync.WaitGroup{}

	for channel := range cchan {
		wg.Add(1)
		go forward(channel, m, wg)
	}

	wg.Wait()
}

func forwardMessage(mchan chan Message, m Message, wg *sync.WaitGroup) {
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

// startgRPCServer starts the message server on a given port
func startCourierServer(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, lis net.Listener, port string) {
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
