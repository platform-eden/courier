package courier

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// TestNodeOptions that can be passed into CreateTestNodes
type TestNodeOptions struct {
	SubscribedSubjects  []string
	BroadcastedSubjects []string
}

// CreateTestNodes creates a quantity of randomized nodes based on the options passed in
func CreateTestNodes(count int, options *TestNodeOptions) []*Node {
	nodes := []*Node{}
	var broadSubjects []string
	var subSubjects []string

	if len(options.SubscribedSubjects) == 0 {
		subSubjects = []string{"sub", "sub1", "sub2"}
	} else {
		subSubjects = options.SubscribedSubjects
	}
	if len(options.BroadcastedSubjects) == 0 {
		broadSubjects = []string{"broad", "broad1"}
	} else {
		broadSubjects = options.BroadcastedSubjects
	}

	for i := 0; i < count; i++ {
		ip := fmt.Sprintf("%v.%v.%v.%v", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
		port := fmt.Sprint(rand.Intn(9999-1000) + 1000)
		subcount := (rand.Intn(len(subSubjects)) + 1)
		broadcount := rand.Intn(len(broadSubjects) + 1)
		var subs []string
		var broads []string

		for i := 0; i < subcount; i++ {
			subs = append(subs, subSubjects[i])
		}

		for i := 0; i < broadcount; i++ {
			broads = append(broads, broadSubjects[i])
		}

		n := NewNode(uuid.NewString(), ip, port, subs, broads)
		nodes = append(nodes, n)
	}

	return nodes
}

func RemovePointers(nodes []*Node) []Node {
	updated := []Node{}

	for _, n := range nodes {
		updated = append(updated, *n)
	}

	return updated
}

type MockServer struct {
	Messages  []*Message
	Responses []*ResponseInfo
	Lis       *bufconn.Listener
	proto.UnimplementedMessageServerServer
	lock       Locker
	shouldFail bool
}

func NewMockServer(listener *bufconn.Listener, fail bool) *MockServer {
	s := MockServer{
		Messages:   []*Message{},
		Responses:  []*ResponseInfo{},
		Lis:        listener,
		lock:       NewTicketLock(),
		shouldFail: fail,
	}

	s.Start()

	return &s
}

func (m *MockServer) Start() {
	grpcServer := grpc.NewServer()

	proto.RegisterMessageServerServer(grpcServer, m)

	go func() {
		err := grpcServer.Serve(m.Lis)
		if err != nil {
			log.Fatalf("server exited with error: %v", err)
		}
	}()
}

func (m *MockServer) BufDialer(context.Context, string) (net.Conn, error) {
	return m.Lis.Dial()
}

func (m *MockServer) PublishMessage(ctx context.Context, request *proto.PublishMessageRequest) (*proto.PublishMessageResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.shouldFail {
		return nil, fmt.Errorf("fail")
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
		return nil, fmt.Errorf("fail")
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
		return nil, fmt.Errorf("fail")
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

func NewLocalGRPCClient(target string, bufDialer func(context.Context, string) (net.Conn, error)) (proto.MessageServerClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	client := proto.NewMessageServerClient(conn)

	return client, conn, nil

}

type MockObserver struct {
	observeChannel chan []Node
	fail           bool
}

func newMockObserver(ochan chan []Node, fail bool) *MockObserver {
	o := MockObserver{
		observeChannel: ochan,
		fail:           fail,
	}

	return &o
}

func (o *MockObserver) Observe() (chan []Node, error) {
	if o.fail {
		return nil, errors.New("fail")
	}

	return o.observeChannel, nil
}

func (o *MockObserver) AddNode(*Node) error {
	if o.fail {
		return errors.New("fail")
	}

	return nil
}

func (o *MockObserver) RemoveNode(*Node) error {
	if o.fail {
		return errors.New("fail")
	}

	return nil
}

type MockClientNode struct {
	Node
	PublishCount  int
	RequestCount  int
	ResponseCount int
	Fail          bool
	Lock          Locker
}

func NewMockClientNode(node Node, fail bool) *MockClientNode {
	c := MockClientNode{
		Node:          node,
		PublishCount:  0,
		RequestCount:  0,
		ResponseCount: 0,
		Fail:          fail,
		Lock:          NewTicketLock(),
	}

	return &c
}

func (c *MockClientNode) sendMessage(ctx context.Context, m Message) error {
	var err error
	switch m.Type {
	case PubMessage:
		err = c.sendPublishMessage(ctx, m)
	case ReqMessage:
		err = c.sendRequestMessage(ctx, m)
	case RespMessage:
		err = c.sendResponseMessage(ctx, m)
	}

	if err != nil {
		return fmt.Errorf("error semding message: %s", err)
	}

	return nil
}

func (c *MockClientNode) sendPublishMessage(ctx context.Context, m Message) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Fail {
		return fmt.Errorf("fail")
	}

	if m.Type != PubMessage {
		return fmt.Errorf("message type must be of type PublishMessage")
	}

	c.PublishCount++

	return nil
}

func (c *MockClientNode) sendRequestMessage(ctx context.Context, m Message) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Fail {
		return fmt.Errorf("fail")
	}

	if m.Type != ReqMessage {
		return fmt.Errorf("message type must be of type RequestMessage")
	}

	c.RequestCount++
	return nil
}

func (c *MockClientNode) sendResponseMessage(ctx context.Context, m Message) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Fail {
		return fmt.Errorf("fail")
	}

	if m.Type != RespMessage {
		return fmt.Errorf("message type must be of type ResponseMessage")
	}

	c.ResponseCount++
	return nil
}

func (c *MockClientNode) Receiver() Node {
	return c.Node
}
