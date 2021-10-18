package mock

import (
	"context"
	"log"
	"net"

	"github.com/platform-edn/courier/lock"
	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type MockServer struct {
	Messages  []*message.Message
	Responses []*node.ResponseInfo
	Lis       *bufconn.Listener
	proto.UnimplementedMessageServerServer
	lock lock.Locker
}

func NewMockServer(listener *bufconn.Listener) *MockServer {
	s := MockServer{
		Messages:  []*message.Message{},
		Responses: []*node.ResponseInfo{},
		Lis:       listener,
		lock:      lock.NewTicketLock(),
	}

	s.Start()

	return &s
}

func (m *MockServer) Start() {
	grpcServer := grpc.NewServer()

	proto.RegisterMessageServerServer(grpcServer, m)

	ok := make(chan bool)
	defer close(ok)
	err := make(chan error)
	defer close(err)

	go func() {
		if e := grpcServer.Serve(m.Lis); err != nil {
			log.Fatalf("server exited with error: %v", e)
		}
	}()
}

func (m *MockServer) BufDialer(context.Context, string) (net.Conn, error) {
	return m.Lis.Dial()
}

func (m *MockServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	pub := message.NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	m.Messages = append(m.Messages, &pub)

	response := proto.PublishMessageResponse{}

	return &response, nil
}

func (m *MockServer) RequestMessage(ctx context.Context, request *proto.RequestMessageRequest) (*proto.RequestMessageResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	req := message.NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := node.ResponseInfo{
		MessageId: request.Message.Id,
		NodeId:    request.Message.NodeId,
	}

	m.Messages = append(m.Messages, &req)
	m.Responses = append(m.Responses, &info)

	response := proto.RequestMessageResponse{}

	return &response, nil
}

func (m *MockServer) ResponseMessage(ctx context.Context, request *proto.ResponseMessageRequest) (*proto.ResponseMessageResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	resp := message.NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	m.Messages = append(m.Messages, &resp)

	response := proto.ResponseMessageResponse{}

	return &response, nil
}

func (m *MockServer) MessagesLength() int {
	m.lock.Lock()
	defer m.lock.Unlock()

	return len(m.Messages)
}

func (m *MockServer) ResponsesLength() int {
	m.lock.Lock()
	defer m.lock.Unlock()

	return len(m.Responses)
}
