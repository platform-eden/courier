package eventer

import (
	"context"
	"fmt"
	"sync"

	"github.com/platform-edn/courier/pkg/proto"
	"github.com/platform-edn/courier/pkg/registry"
	"google.golang.org/grpc"
)

type EventStreamer interface {
	StreamEvents(ctx context.Context, subscribed []string, broadcasted []string) error
}

type OperatorClient struct {
	EventOut          chan registry.NodeEvent
	OperatorAddress   string
	OperatorPort      string
	Subscribed        []string
	Broadcasted       []string
	ConnectionOptions []grpc.DialOption
	EventStreamer
}

type OperatorClientOption func(client *OperatorClient)

func WithEventChannel(out chan registry.NodeEvent) OperatorClientOption {
	return func(client *OperatorClient) {
		client.EventOut = out
	}
}

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

func WithBroadcasted(subjects ...string) OperatorClientOption {
	return func(client *OperatorClient) {
		client.Broadcasted = append(client.Broadcasted, subjects...)
	}
}

func WithSubscribed(subjects ...string) OperatorClientOption {
	return func(client *OperatorClient) {
		client.Subscribed = append(client.Subscribed, subjects...)
	}
}

func NewOperatorClient(options ...OperatorClientOption) (*OperatorClient, error) {
	client := &OperatorClient{}

	for _, option := range options {
		option(client)
	}

	var err error
	switch {
	case client.EventOut == nil:
		err = &MissingOperatorClientParamError{
			Param: "EventOut",
		}
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
	case len(client.Broadcasted) == 0 && len(client.Subscribed) == 0:
		err = &MissingOperatorClientParamError{
			Param: "Broadcasted and Subscribed Subjects",
		}
	default:
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("NewOperatorClient: %w", err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", client.OperatorAddress, client.OperatorPort), client.ConnectionOptions...)
	if err != nil {
		return nil, fmt.Errorf("NewOperatorClient: %w", err)
	}

	client.EventStreamer = NewEventStream(client.EventOut, proto.NewDiscoverClient(conn))

	return client, nil
}

func (client *OperatorClient) DiscoverNodeEvents(ctx context.Context, out chan registry.NodeEvent, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := client.StreamEvents(ctx, client.Subscribed, client.Broadcasted)
			if err != nil {
				errs <- err
			}
		}
	}
}
