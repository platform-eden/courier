package mocks

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/message"
)

func TestNewMockSender(t *testing.T) {
	s := NewMockSender()

	if s.PublishCount+s.RequestCount+s.ResponseCount+s.Length() != 0 {
		t.Fatalf("expected MockSender to have zero values for properties but they were not")
	}
}

func TestMockSender_Send(t *testing.T) {
	s := NewMockSender()
	ctx := context.Background()

	pub := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	req := message.NewReqMessage(uuid.NewString(), "test", []byte("test"))
	resp := message.NewRespMessage(uuid.NewString(), "test", []byte("test"))

	err := s.Send(ctx, pub, nil)
	if err != nil {
		t.Fatalf("expected send to pass but it didn't: %s", err)
	}
	if s.PublishCount != 1 || s.Length() != 1 {
		t.Fatalf("incorrect quantity of messages after sending PublishMessage - count: %v, length: %v", s.PublishCount, s.Length())
	}
	err = s.Send(ctx, req, nil)
	if err != nil {
		t.Fatalf("expected send to pass but it didn't: %s", err)
	}
	if s.RequestCount != 1 || s.Length() != 2 {
		t.Fatalf("incorrect quantity of messages after sending RequestMessage - count: %v, length: %v", s.RequestCount, s.Length())
	}
	err = s.Send(ctx, resp, nil)
	if err != nil {
		t.Fatalf("expected send to pass but it didn't: %s", err)
	}
	if s.ResponseCount != 1 || s.Length() != 3 {
		t.Fatalf("incorrect quantity of messages after sending ResponseMessage - count: %v, length: %v", s.ResponseCount, s.Length())
	}

	s.Fail = true

	err = s.Send(ctx, pub, nil)
	if err == nil {
		t.Fatal("expected send to fail but it passed")
	}

}
