package courier

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type MessageServer struct {
	responseChannel chan ResponseInfo
	channelMap      channelMapper
	proto.UnimplementedMessageServerServer
}

func NewMessageServer(rchan chan ResponseInfo, chanMap channelMapper) *MessageServer {
	m := MessageServer{
		responseChannel: rchan,
		channelMap:      chanMap,
	}

	return &m
}

func (m *MessageServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	pub := NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

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
	req := NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := ResponseInfo{
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
	resp := NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	channels, err := m.channelMap.Subscriptions(resp.Subject)
	if err != nil {
		return nil, fmt.Errorf("could not get subscriptions for %s: %s", resp.Subject, err)
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

func getListener(port string) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("could not listen on port %s: %s", port, err)
	}

	return lis, nil
}

// startMessageServer starts the message server on a given port
func startMessageServer(m proto.MessageServerServer, lis net.Listener) {
	grpcServer := grpc.NewServer()
	proto.RegisterMessageServerServer(grpcServer, m)
	err := grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("error serving grpc: %s", err)
	}
}
