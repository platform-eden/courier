package eventer

import (
	"context"
	"fmt"
	"sync"

	"github.com/platform-edn/courier/pkg/proto"
	"github.com/platform-edn/courier/pkg/registry"
	"google.golang.org/grpc"
)

type OperatorClient struct {
	OperatorAddress   string
	OperatorPort      string
	Subscribed        []string
	Broadcasted       []string
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

	client.DiscoverClient = proto.NewDiscoverClient(conn)

	return client, nil
}

// DiscoverNodeEvents attempts to create a stream with the discovery server and forwards node events in to the system.  Will always
// reattempt connection until context to close is callled.
func (client *OperatorClient) DiscoverNodeEvents(ctx context.Context, out chan registry.NodeEvent, errs chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	eventTypes := NewEventTypeMap()

	for {
		select {
		// make sure we should still be running
		case <-ctx.Done():
			return
		// always attempt to get stream if we are supposed to be running
		default:
			stream, err := client.SubscribeToNodeEvents(ctx, &proto.EventStreamRequest{
				SubscribedSubjects:  client.Subscribed,
				BroadcastedSubjects: client.Broadcasted,
			})
			//break from select and loop to context check
			if err != nil {
				errs <- err
				break
			}

			for {
				resp, err := stream.Recv()
				//break from for and loop to context check
				if err != nil {
					errs <- err
					break
				}

				respNode := resp.Event.Node
				out <- registry.NewNodeEvent(
					*registry.NewNode(respNode.Id, respNode.Address, respNode.Port, respNode.SubscribedSubjects, respNode.BroadcastedSubjects),
					eventTypes[resp.Event.EventType],
				)
			}
		}
	}
}
