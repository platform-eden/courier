package mock

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc/test/bufconn"
)

func TestMockServer_Start(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := NewMockServer(lis)

	_, _, err := NewLocalGRPCClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
}

func TestMockServer_PublishMessage(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := NewMockServer(lis)
	ctx := context.Background()

	client, conn, err := NewLocalGRPCClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
	defer conn.Close()

	m := proto.PublishMessage{
		Id:      uuid.NewString(),
		Subject: "test",
		Content: []byte("test"),
	}

	_, err = client.PublishMessage(ctx, &proto.PublishMessageRequest{Message: &m})
	if err != nil {
		t.Fatalf("error sending publish message: %s", err)
	}

	if s.MessagesLength() != 1 {
		t.Fatalf("incorrect message slice size of %v", s.MessagesLength())
	}

}

func TestMockServer_ReqMessage(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := NewMockServer(lis)
	ctx := context.Background()

	client, conn, err := NewLocalGRPCClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
	defer conn.Close()

	m := proto.RequestMessage{
		Id:      uuid.NewString(),
		NodeId:  "test.com",
		Subject: "test",
		Content: []byte("test"),
	}

	_, err = client.RequestMessage(ctx, &proto.RequestMessageRequest{Message: &m})
	if err != nil {
		t.Fatalf("error sending publish message: %s", err)
	}

	if s.MessagesLength() != 1 {
		t.Fatalf("incorrect message slice size of %v", s.MessagesLength())
	}

	if s.ResponsesLength() != 1 {
		t.Fatalf("incorrect response slice size of %v", s.ResponsesLength())
	}

}

func TestMockServer_RespMessage(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := NewMockServer(lis)
	ctx := context.Background()

	client, conn, err := NewLocalGRPCClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
	defer conn.Close()

	m := proto.ResponseMessage{
		Id:      uuid.NewString(),
		Subject: "test",
		Content: []byte("test"),
	}

	_, err = client.ResponseMessage(ctx, &proto.ResponseMessageRequest{Message: &m})
	if err != nil {
		t.Fatalf("error sending publish message: %s", err)
	}

	if s.MessagesLength() != 1 {
		t.Fatalf("incorrect message slice size of %v", s.MessagesLength())
	}

}
