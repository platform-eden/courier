package messaging

import "context"

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
		return &ClientNodeMessageTypeError{
			Type:   ReqMessage,
			Method: "sendRequestMessage",
		}
	}

	return nil
}

func (c *MockClientNode) sendPublishMessage(ctx context.Context, m Message) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Fail {
		return &ExpectedFailureError{}
	}

	if m.Type != PubMessage {
		return &ClientNodeMessageTypeError{
			Type:   ReqMessage,
			Method: "sendPublishMessage",
		}
	}

	c.PublishCount++

	return nil
}

func (c *MockClientNode) sendRequestMessage(ctx context.Context, m Message) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Fail {
		return &ExpectedFailureError{}
	}

	if m.Type != ReqMessage {
		return &ClientNodeMessageTypeError{
			Type:   ReqMessage,
			Method: "sendRequestMessage",
		}
	}

	c.RequestCount++
	return nil
}

func (c *MockClientNode) sendResponseMessage(ctx context.Context, m Message) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Fail {
		return &ExpectedFailureError{}
	}

	if m.Type != RespMessage {
		return &ClientNodeMessageTypeError{
			Type:   ReqMessage,
			Method: "sendResponseMessage",
		}
	}

	c.ResponseCount++
	return nil
}

func (c *MockClientNode) Receiver() Node {
	return c.Node
}
