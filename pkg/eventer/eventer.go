package eventer

import (
	"context"
	"fmt"

	"github.com/platform-edn/courier/pkg/proto"
	"github.com/platform-edn/courier/pkg/registry"
	"google.golang.org/grpc"
)

type OperatorClient struct {
	OperatorAddress   string
	OperatorPort      string
	ConnectionOptions []grpc.DialOption
	proto.DiscoverClient
}

type OperatorClientOption func(client *OperatorClient)

func WithOperatorAddress(address string) OperatorClientOption {
	return func(client *OperatorClient) {
		client.OperatorAddress = address
	}
}

func WithOperatorPort(port string) OperatorClientOption {
	return func(client *OperatorClient) {
		client.OperatorPort = port
	}
}

func WithConnectionOptions(options ...grpc.DialOption) OperatorClientOption {
	return func(client *OperatorClient) {
		client.ConnectionOptions = options
	}
}

func NewOperatorClient(options ...OperatorClientOption) (*OperatorClient, error) {
	client := &OperatorClient{}

	for _, option := range options {
		option(client)
	}

	var err error
	switch {
	case client.OperatorAddress == "":
		err = &MissingOperatorClientParamError{
			Param: "OperatorAddress",
		}
	case client.OperatorPort == "":
		err = &MissingOperatorClientParamError{
			Param: "OperatorPort",
		}
	case len(client.ConnectionOptions) == 0:
		err = &MissingOperatorClientParamError{
			Param: "ConnectionOptions",
		}
	default:
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("NewOperatorClient: %w", err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", client.OperatorAddress, client.OperatorPort), client.ConnectionOptions)
	if err != nil {
		return nil, fmt.Errorf("NewOperatorClient: %w", err)
	}

	client.DiscoverClient = proto.NewDiscoverClient(conn)

	return client, nil
}

func (client *OperatorClient) DiscoverNodeEvents(ctx context.Context, out chan registry.NodeEvent) {}
