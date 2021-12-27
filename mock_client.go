package courier

import (
	"context"
	"fmt"
	"net"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

func NewMockClient(target string, bufDialer func(context.Context, string) (net.Conn, error), options ...grpc.DialOption) (proto.MessageServerClient, *grpc.ClientConn, error) {
	mockDialOptions := []grpc.DialOption{
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure(),
	}

	for _, option := range options {
		mockDialOptions = append(mockDialOptions, option)
	}

	conn, err := grpc.DialContext(context.Background(), target, mockDialOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	client := proto.NewMessageServerClient(conn)

	return client, conn, nil

}
