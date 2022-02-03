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
	PopResponse(string) (string, error)
	Length() int
}

type SubMapper interface {
	AddSubscriber(string, ...string)
	RemoveSubscriber(string, ...string)
	Subscribers(string) ([]string, error)
	CheckForSubscriber(subject string, id string) bool
}

type ClientNodeMapper interface {
	Node(string) (clientNode, bool)
	AddClientNode(clientNode)
	RemoveClientNode(string)
	Length() int
}

type Sender interface {
	sendMessage(context.Context, messaging.Message) error
	Receiver() registry.Node
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
		ClientNodeMapper: newClientNodeMap(),
		SubMapper:        newSubscriberMap(),
		ResponseMapper:   newResponseMap(),
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
				cn, err := newClientNode(e.Node, currentId, client.clientOptions...)
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
	ids, err := generateIdsBySubject(msg.Subject, c.SubMapper)
	if err != nil {
		return &MessagingClientError{
			Method: "Publish",
			Err:    err,
		}
	}

	cnodes := idToClientNodes(ids, c.ClientNodeMapper)
	failed := fanMessageAttempts(cnodes, ctx, msg)
	done := forwardFailedConnections(failed, c.ClientNodeMapper, c.SubMapper, c.failedEvents)

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
	ids, err := generateIdsByMessage(msg.Id, c.ResponseMapper)
	if err != nil {
		return &MessagingClientError{
			Method: "Response",
			Err:    err,
		}
	}

	cnodes := idToClientNodes(ids, c.ClientNodeMapper)
	failed := fanMessageAttempts(cnodes, ctx, msg)
	done := forwardFailedConnections(failed, c.ClientNodeMapper, c.SubMapper, c.failedEvents)

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

func (c *messagingClient) Request(ctx context.Context, msg messaging.Message) error {
	ids, err := generateIdsBySubject(msg.Subject, c.SubMapper)
	if err != nil {
		return &MessagingClientError{
			Method: "Request",
			Err:    err,
		}
	}

	cnodes := idToClientNodes(ids, c.ClientNodeMapper)
	failed := fanMessageAttempts(cnodes, ctx, msg)
	done := forwardFailedConnections(failed, c.ClientNodeMapper, c.SubMapper, c.failedEvents)

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

// generateIdsBySubject takes a subject string and returns a channel of node ids subscribing to it
func generateIdsBySubject(subject string, subMap SubMapper) (<-chan string, error) {
	out := make(chan string)

	ids, err := subMap.Subscribers(subject)
	if err != nil {
		return nil, &NodeIdGenerationError{
			Method: "generateIdsBySubject",
			Err:    err,
		}
	}

	go func() {
		for _, id := range ids {
			out <- id
		}
		close(out)
	}()

	return out, nil
}

// generateIdsByMessage takes a messageId and returns a channel of node ids expecting to receive a response
func generateIdsByMessage(messageId string, respMap ResponseMapper) (<-chan string, error) {
	out := make(chan string, 1)

	id, err := respMap.PopResponse(messageId)
	if err != nil {
		return nil, &NodeIdGenerationError{
			Method: "generateIdsByMessage",
			Err:    err,
		}
	}

	out <- id
	close(out)

	return out, nil
}

// idToNodes takes a channel of ids and returns a channel of nodes based on the ids
func idToClientNodes(in <-chan string, nodeMap ClientNodeMapper) <-chan Sender {
	out := make(chan Sender)
	go func() {
		for id := range in {
			n, exist := nodeMap.Node(id)
			if !exist {
				log.Printf("node %s does not exist in nodemap - skipping", id)
				continue
			}

			out <- &n
		}
		close(out)
	}()

	return out
}

// fanMessageAttempts takes a channel of Senders and creates a goroutine for each to attempt to send a message. Returns a channel that will return nodes that unsuccessfully sent a message
func fanMessageAttempts(in <-chan Sender, ctx context.Context, msg messaging.Message) chan registry.Node {
	out := make(chan registry.Node)

	go func() {
		wg := &sync.WaitGroup{}

		for n := range in {
			wg.Add(1)
			go attemptMessage(ctx, n, msg, out, wg)
		}

		wg.Wait()
		close(out)
	}()

	return out
}

// attemptMessage takes a send function and attempts it until it succeeds or has reached maxAttempts.  Waits between attempts depends on the given interval
func attemptMessage(ctx context.Context, sender Sender, msg messaging.Message, nchan chan registry.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	err := sender.sendMessage(ctx, msg)
	if err != nil {
		nchan <- sender.Receiver()
	}
}

// forwardFailedConnections takes a channel of nodes and sends them through a stale channel as well as a failed connection channel.  Returns a bool channel that receives true when it's done
func forwardFailedConnections(in <-chan registry.Node, clientNodes ClientNodeMapper, subscribers SubMapper, fchan chan registry.NodeEvent) <-chan bool {
	out := make(chan bool)

	go func() {
		for n := range in {
			log.Printf("failed creating connection to %s:%s - will now be removing and blacklisting %s\n", n.Address, n.Port, n.Id)
			clientNodes.RemoveClientNode(n.Id)
			subscribers.RemoveSubscriber(n.Id, n.SubscribedSubjects...)

			event := registry.NewNodeEvent(n, registry.Failed)
			fchan <- *event

		}
		out <- true
		close(out)
	}()

	return out
}
