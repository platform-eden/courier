package courier

import (
	"context"
	"fmt"
	"net"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

func NewMockClient(target string, bufDialer func(context.Context, string) (net.Conn, error)) (proto.MessageServerClient, *grpc.ClientConn, error) {
	conn, err := grpc.DialContext(context.Background(), target, grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial bufnet: %v", err)
	}

	client := proto.NewMessageServerClient(conn)

	return client, conn, nil

}
