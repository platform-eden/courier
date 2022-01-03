package messaging

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/platform-edn/courier/proto"
// )

// func TestMockServer_PublishMessage(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
// 	defer cancel()

// 	client, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
// 	if err != nil {
// 		t.Fatalf("could not creat client for server: %s", err)
// 	}
// 	defer conn.Close()

// 	m := proto.PublishMessage{
// 		Id:      uuid.NewString(),
// 		Subject: "test",
// 		Content: []byte("test"),
// 	}

// 	_, err = client.PublishMessage(ctx, &proto.PublishMessageRequest{Message: &m})
// 	if err != nil {
// 		t.Fatalf("error sending publish message: %s", err)
// 	}

// 	if testMessageServer.MessagesLength() != 1 {
// 		t.Fatalf("incorrect message slice size of %v", testMessageServer.MessagesLength())
// 	}

// }

// func TestMockServer_ReqMessage(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
// 	defer cancel()

// 	client, conn, err := NewMockClient("bufnset", testMessageServer.BufDialer)
// 	if err != nil {
// 		t.Fatalf("could not creat client for server: %s", err)
// 	}
// 	defer conn.Close()

// 	m := proto.RequestMessage{
// 		Id:      uuid.NewString(),
// 		NodeId:  "test.com",
// 		Subject: "test",
// 		Content: []byte("test"),
// 	}

// 	_, err = client.RequestMessage(ctx, &proto.RequestMessageRequest{Message: &m})
// 	if err != nil {
// 		t.Fatalf("error sending publish message: %s", err)
// 	}

// 	if testMessageServer.MessagesLength() != 1 {
// 		t.Fatalf("incorrect message slice size of %v", testMessageServer.MessagesLength())
// 	}

// 	if testMessageServer.ResponsesLength() != 1 {
// 		t.Fatalf("incorrect response slice size of %v", testMessageServer.ResponsesLength())
// 	}

// }

// func TestMockServer_RespMessage(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*3))
// 	defer cancel()

// 	client, conn, err := NewMockClient("bufnet", testMessageServer.BufDialer)
// 	if err != nil {
// 		t.Fatalf("could not creat client for server: %s", err)
// 	}
// 	defer conn.Close()

// 	m := proto.ResponseMessage{
// 		Id:      uuid.NewString(),
// 		Subject: "test",
// 		Content: []byte("test"),
// 	}

// 	_, err = client.ResponseMessage(ctx, &proto.ResponseMessageRequest{Message: &m})
// 	if err != nil {
// 		t.Fatalf("error sending publish message: %s", err)
// 	}

// 	if testMessageServer.MessagesLength() != 1 {
// 		t.Fatalf("incorrect message slice size of %v", testMessageServer.MessagesLength())
// 	}

// }
