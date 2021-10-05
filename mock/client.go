package mock

import (
	"context"
	"fmt"
	"net"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

func NewMockClient(ctx context.Context, target string, bufDialer func(context.Context, string) (net.Conn, error)) (proto.MessageServerClient, *grpc.ClientConn, error) {
	conn, err := grpc.DialContext(ctx, target, grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	client := proto.NewMessageServerClient(conn)

	return client, conn, nil

}
