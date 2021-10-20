package client

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/mock"
	"google.golang.org/grpc/test/bufconn"
)

func TestPublishClient_Send(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := mock.NewMockServer(lis)

	client, conn, err := mock.NewMockClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
	defer conn.Close()
	m := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	m1 := message.NewReqMessage(uuid.NewString(), "test", []byte("test"))
	c := PublishClient{}

	err = c.Send(context.Background(), m, client)
	if err != nil {
		t.Fatalf("could not send message: %s", err)
	}
	err = c.Send(context.Background(), m1, client)
	if err == nil {
		t.Fatal("expected message to fail but it passed")
	}
}

func TestRequestClient_Send(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := mock.NewMockServer(lis)

	client, conn, err := mock.NewMockClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
	defer conn.Close()
	m := message.NewReqMessage(uuid.NewString(), "test", []byte("test"))
	m1 := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	c := RequestClient{NodeId: uuid.NewString()}

	err = c.Send(context.Background(), m, client)
	if err != nil {
		t.Fatalf("could not send message: %s", err)
	}
	err = c.Send(context.Background(), m1, client)
	if err == nil {
		t.Fatal("expected message to fail but it passed")
	}
}

func TestResponseClient_Send(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := mock.NewMockServer(lis)

	client, conn, err := mock.NewMockClient("bufnet", s.BufDialer)
	if err != nil {
		t.Fatalf("could not creat client for server: %s", err)
	}
	defer conn.Close()
	m := message.NewRespMessage(uuid.NewString(), "test", []byte("test"))
	m1 := message.NewPubMessage(uuid.NewString(), "test", []byte("test"))
	c := ResponseClient{}

	err = c.Send(context.Background(), m, client)
	if err != nil {
		t.Fatalf("could not send message: %s", err)
	}
	err = c.Send(context.Background(), m1, client)
	if err == nil {
		t.Fatal("expected message to fail but it passed")
	}
}
