package client

import (
	"context"
	"fmt"
	"net"

	"github.com/platform-edn/courier/pkg/messaging/proto"
	"google.golang.org/grpc"
)

func NewMockClient(target string, bufDialer func(context.Context, string) (net.Conn, error), options ...grpc.DialOption) (proto.MessageClient, *grpc.ClientConn, error) {
	mockDialOptions := []grpc.DialOption{
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure(),
	}
	mockDialOptions = append(mockDialOptions, options...)

	conn, err := grpc.DialContext(context.Background(), target, mockDialOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	client := proto.NewMessageClient(conn)

	return client, conn, nil

}
