package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/registry"
)

type ResponseMapper interface {
	PushResponse(messaging.ResponseInfo)
	GenerateIdsByMessage(string) (<-chan string, error)
}

type SubMapper interface {
	AddSubscriber(string, ...string)
	RemoveSubscriber(string, ...string)
	Subscribers(string) ([]string, error)
	GenerateIdsBySubject(string) (<-chan string, error)
}

type ClientNodeMapper interface {
	Node(string) (*ClientNode, error)
	AddClientNode(ClientNode)
	RemoveClientNode(string)
	FanClientNodeMessaging(context.Context, messaging.Message, <-chan string) chan registry.Node
}
type messagingClient struct {
	ClientNodeMapper
	SubMapper
	ResponseMapper
	clientOptions []ClientNodeOption
	failedEvents  chan registry.NodeEvent
}

func NewMessagingClient(failedChannel chan registry.NodeEvent, options ...ClientNodeOption) *messagingClient {
	c := messagingClient{
		ClientNodeMapper: NewClientNodeMap(),
		SubMapper:        NewSubscriberMap(),
		ResponseMapper:   NewResponseMap(),
		clientOptions:    options,
		failedEvents:     failedChannel,
	}

	return &c
}

func (c *messagingClient) ListenForResponseInfo(ctx context.Context, wg *sync.WaitGroup, responses <-chan messaging.ResponseInfo) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case response := <-responses:
			c.PushResponse(response)
		}
	}
}

// listenForNewNodes takes nodes passed through a channel and adds them to a NodeMapper and SubMapper
func (client *messagingClient) ListenForNodeEvents(ctx context.Context, wg *sync.WaitGroup, events <-chan registry.NodeEvent, errChan chan error, currentId string) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-events:
			switch e.Event {
			case registry.Add:
				cn, err := NewClientNode(e.Node, currentId, client.clientOptions...)
				if err != nil {
					log.Printf("skipping %s, couldn't create client node: %s", e.Node.Id, err)
					continue
				}
				client.AddClientNode(*cn)
			case registry.Remove:
				client.RemoveClientNode(e.Id)
				client.RemoveSubscriber(e.Id, e.SubscribedSubjects...)
			default:
				errChan <- fmt.Errorf("ListenForNodeEvents: %w", &UnknownNodeEventError{
					Event: e.Event.String(),
					Id:    e.Id,
				})
			}
		}
	}
}

func (c *messagingClient) Publish(ctx context.Context, msg messaging.Message) error {
	if msg.Type != messaging.PubMessage {
		return fmt.Errorf("Publish: %w", &BadMessageTypeError{
			Expected: messaging.PubMessage.String(),
			Actual:   msg.Type.String(),
		})
	}

	ids, err := c.GenerateIdsBySubject(msg.Subject)
	if err != nil {
		return fmt.Errorf("Publish: %w", err)
	}

	failed := c.FanClientNodeMessaging(ctx, msg, ids)
	done := c.ForwardFailedConnections(failed)

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("Publish: %w", &ContextDoneUnsentMessageError{
			MessageId: msg.Id,
		})
	}

	return nil
}

func (c *messagingClient) Response(ctx context.Context, msg messaging.Message) error {
	if msg.Type != messaging.RespMessage {
		return fmt.Errorf("Response: %w", &BadMessageTypeError{
			Expected: messaging.RespMessage.String(),
			Actual:   msg.Type.String(),
		})
	}

	ids, err := c.GenerateIdsByMessage(msg.Id)
	if err != nil {
		return fmt.Errorf("Response: %w", err)
	}

	failed := c.FanClientNodeMessaging(ctx, msg, ids)
	done := c.ForwardFailedConnections(failed)

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("Response: %w", &ContextDoneUnsentMessageError{
			MessageId: msg.Id,
		})
	}

	return nil
}

func (c *messagingClient) Request(ctx context.Context, msg messaging.Message) error {
	if msg.Type != messaging.ReqMessage {
		return fmt.Errorf("Request: %w", &BadMessageTypeError{
			Expected: messaging.ReqMessage.String(),
			Actual:   msg.Type.String(),
		})
	}

	ids, err := c.GenerateIdsBySubject(msg.Subject)
	if err != nil {
		return fmt.Errorf("Request: %w", err)
	}

	failed := c.FanClientNodeMessaging(ctx, msg, ids)
	done := c.ForwardFailedConnections(failed)

	select {
	case <-done:
		break
	case <-ctx.Done():
		return fmt.Errorf("Request: %w", &ContextDoneUnsentMessageError{
			MessageId: msg.Id,
		})
	}

	return nil
}

// forwardFailedConnections takes a channel of nodes and sends them through a stale channel as well as a failed connection channel.  Returns a bool channel that receives true when it's done
func (c *messagingClient) ForwardFailedConnections(in chan registry.Node) <-chan struct{} {
	out := make(chan struct{})

	go func() {
		for n := range in {
			log.Printf("failed creating connection to %s:%s - will now be removing and blacklisting %s\n", n.Address, n.Port, n.Id)
			c.RemoveClientNode(n.Id)
			c.RemoveSubscriber(n.Id, n.SubscribedSubjects...)

			event := registry.NewNodeEvent(n, registry.Failed)
			c.failedEvents <- *event

		}
		close(out)
	}()

	return out
}
