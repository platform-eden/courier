package messaging

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type ResponseMapper interface {
	Push(ResponseInfo)
	Pop(string) (string, error)
	Length() int
}

type SubMapper interface {
	Add(string, ...string)
	Remove(string, ...string)
	Subscribers(string) ([]string, error)
	CheckForSubscriber(subject string, id string) bool
}

type ClientNodeMapper interface {
	Node(string) (clientNode, bool)
	Add(clientNode)
	Remove(string)
	Length() int
}

type Sender interface {
	sendMessage(context.Context, Message) error
	Receiver() Node
}

type messagingClient struct {
	clientNodes     ClientNodeMapper
	subscribers     SubMapper
	responses       ResponseMapper
	failedChannel   chan Node
	staleChannel    chan Node
	newNodeChannel  chan Node
	responseChannel chan ResponseInfo
	currentId       string
	cancelFunc      context.CancelFunc
	waitGroup       *sync.WaitGroup
	clientOptions   []ClientNodeOption
}

type messageClientOptions struct {
	failedChannel   chan Node
	staleChannel    chan Node
	nodeChannel     chan Node
	responseChannel chan ResponseInfo
	currentId       string
	clientOptions   []ClientNodeOption
	startClient     bool
}

type MessagingClientError struct {
	Method string
	Err    error
}

func (err *MessagingClientError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type NodeIdGenerationError struct {
	Method string
	Err    error
}

func (err *NodeIdGenerationError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

func newMessagingClient(options *messageClientOptions) *messagingClient {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(3)

	c := messagingClient{
		clientNodes:     newClientNodeMap(),
		subscribers:     newSubscriberMap(),
		responses:       newResponseMap(),
		failedChannel:   options.failedChannel,
		staleChannel:    options.staleChannel,
		responseChannel: options.responseChannel,
		newNodeChannel:  options.nodeChannel,
		currentId:       options.currentId,
		clientOptions:   options.clientOptions,
		cancelFunc:      cancel,
		waitGroup:       wg,
	}

	if len(c.clientOptions) == 0 {
		c.clientOptions = append(c.clientOptions, WithInsecure())
	}
	if options.startClient {
		go c.listenForNewNodes(ctx, wg)
		go c.listenForResponseInfo(ctx, wg)
		go c.listenForStaleNodes(ctx, wg)
	}

	return &c
}

func (c *messagingClient) stop() {
	c.cancelFunc()
	c.waitGroup.Wait()

	// close channels we write on
	close(c.failedChannel)
}

func (c *messagingClient) listenForResponseInfo(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case response := <-c.responseChannel:
			c.responses.Push(response)
		}
	}
}

// listenForNewNodes takes nodes passed through a channel and adds them to a NodeMapper and SubMapper
func (c *messagingClient) listenForNewNodes(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-c.newNodeChannel:
			cn, err := newClientNode(n, c.currentId, c.clientOptions...)
			if err != nil {
				log.Printf("skipping %s, couldn't create client node: %s", n.id, err)
				continue
			}

			c.clientNodes.Add(*cn)
			c.subscribers.Add(n.id, n.subscribedSubjects...)
		}
	}
}

// listenForStaleNodes takes nodes passed in and removes them from a NodeMapper and SubMapper
func (c *messagingClient) listenForStaleNodes(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-c.staleChannel:
			c.clientNodes.Remove(n.id)
			c.subscribers.Remove(n.id, n.subscribedSubjects...)
		}
	}
}

func (c *messagingClient) publish(ctx context.Context, msg Message) error {
	ids, err := generateIdsBySubject(msg.Subject, c.subscribers)
	if err != nil {
		return &MessagingClientError{
			Method: "Publish",
			Err:    err,
		}
	}

	cnodes := idToClientNodes(ids, c.clientNodes)
	failed := fanMessageAttempts(cnodes, ctx, msg)
	done := forwardFailedConnections(failed, c.clientNodes, c.subscribers, c.failedChannel)

	<-done

	return nil
}

func (c *messagingClient) response(ctx context.Context, msg Message) error {
	ids, err := generateIdsByMessage(msg.Id, c.responses)
	if err != nil {
		return &MessagingClientError{
			Method: "Response",
			Err:    err,
		}
	}

	cnodes := idToClientNodes(ids, c.clientNodes)
	failed := fanMessageAttempts(cnodes, ctx, msg)
	done := forwardFailedConnections(failed, c.clientNodes, c.subscribers, c.failedChannel)

	<-done

	return nil
}

func (c *messagingClient) request(ctx context.Context, msg Message) error {
	ids, err := generateIdsBySubject(msg.Subject, c.subscribers)
	if err != nil {
		return &MessagingClientError{
			Method: "Request",
			Err:    err,
		}
	}

	cnodes := idToClientNodes(ids, c.clientNodes)
	failed := fanMessageAttempts(cnodes, ctx, msg)
	done := forwardFailedConnections(failed, c.clientNodes, c.subscribers, c.failedChannel)

	<-done
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

	id, err := respMap.Pop(messageId)
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
func fanMessageAttempts(in <-chan Sender, ctx context.Context, msg Message) chan Node {
	out := make(chan Node)

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
func attemptMessage(ctx context.Context, sender Sender, msg Message, nchan chan Node, wg *sync.WaitGroup) {
	defer wg.Done()
	err := sender.sendMessage(ctx, msg)
	if err != nil {
		nchan <- sender.Receiver()
	}
}

// forwardFailedConnections takes a channel of nodes and sends them through a stale channel as well as a failed connection channel.  Returns a bool channel that receives true when it's done
func forwardFailedConnections(in <-chan Node, clientNodes ClientNodeMapper, subscribers SubMapper, fchan chan Node) <-chan bool {
	out := make(chan bool)

	go func() {
		for n := range in {
			log.Printf("failed creating connection to %s:%s - will now be removing and blacklisting %s\n", n.address, n.port, n.id)
			clientNodes.Remove(n.id)
			subscribers.Remove(n.id, n.subscribedSubjects...)

			fchan <- n

		}
		out <- true
		close(out)
	}()

	return out
}
