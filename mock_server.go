package courier

import (
	"context"
	"net"
	"sync"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type MockServer struct {
	Messages  []*Message
	Responses []*ResponseInfo
	Lis       *bufconn.Listener
	port      string
	proto.UnimplementedMessageServerServer
	lock       Locker
	shouldFail bool
	isRunning  bool
}

func NewMockServer(lis *bufconn.Listener, port string, fail bool) *MockServer {
	s := MockServer{
		Messages:   []*Message{},
		Responses:  []*ResponseInfo{},
		Lis:        lis,
		port:       port,
		lock:       NewTicketLock(),
		shouldFail: fail,
		isRunning:  false,
	}

	return &s
}

func (m *MockServer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	if m.isRunning {
		return nil
	}

	server := grpc.NewServer()
	proto.RegisterMessageServerServer(server, m)

	go startCourierServer(ctx, wg, server, m.Lis, m.port)

	m.isRunning = true
	return nil
}

func (m *MockServer) SetToFail() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.shouldFail = true
}

func (m *MockServer) SetToPass() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.shouldFail = false
}

func (m *MockServer) BufDialer(context.Context, string) (net.Conn, error) {
	return m.Lis.Dial()
}

func (m *MockServer) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.Messages = []*Message{}
	m.Responses = []*ResponseInfo{}
}

func (m *MockServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.shouldFail {
		return nil, &ExpectedFailureError{}
	}

	pub := NewPubMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

	m.Messages = append(m.Messages, &pub)

	response := proto.PublishMessageResponse{}

	return &response, nil
}

func (m *MockServer) RequestMessage(ctx context.Context, request *proto.RequestMessageRequest) (*proto.RequestMessageResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.shouldFail {
		return nil, &ExpectedFailureError{}
	}

	req := NewReqMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())
	info := ResponseInfo{
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
	if m.shouldFail {
		return nil, &ExpectedFailureError{}
	}

	resp := NewRespMessage(request.Message.Id, request.Message.Subject, request.Message.GetContent())

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
