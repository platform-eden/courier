package eventer

import (
	"context"
	"fmt"

	"github.com/platform-edn/courier/pkg/proto"
	"github.com/platform-edn/courier/pkg/registry"
)

type EventStream struct {
	Out chan registry.NodeEvent
	proto.DiscoverClient
}

func NewEventStream(out chan registry.NodeEvent, disco proto.DiscoverClient) *EventStream {
	return &EventStream{
		Out:            out,
		DiscoverClient: disco,
	}
}

func (stream *EventStream) StreamEvents(ctx context.Context, subscribed []string, broadcasted []string) error {
	eventTypes := NewEventTypeMap()
	streamClient, err := stream.SubscribeToNodeEvents(ctx, &proto.EventStreamRequest{
		SubscribedSubjects:  subscribed,
		BroadcastedSubjects: broadcasted,
	})
	if err != nil {
		return fmt.Errorf("StreamEvents: %w", err)
	}

	for {
		resp, err := streamClient.Recv()
		if err != nil {
			return fmt.Errorf("StreamEvents: %w", err)
		}

		respNode := resp.Event.Node
		stream.Out <- registry.NewNodeEvent(
			*registry.NewNode(respNode.Id, respNode.Address, respNode.Port, respNode.SubscribedSubjects, respNode.BroadcastedSubjects),
			eventTypes[resp.Event.EventType],
		)
	}
}
