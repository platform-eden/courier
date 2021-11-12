package courier

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type courierClient struct {
	client     proto.MessageServerClient
	connection grpc.ClientConn
	currentId  string
	receiver   Node
}

type attemptMetadata struct {
	maxAttempts  int
	waitInterval time.Duration
}

// listenForResponseInfo takes ResponseInfo through a channel and pushes them into a responseMap
func listenForResponseInfo(responses chan ResponseInfo, respMap ResponseMapper) {
	for response := range responses {
		respMap.Push(response)
	}
}

// listenForNewNodes takes nodes passed through a channel and adds them to a NodeMapper and SubMapper
func listenForNewNodes(nodeChannel chan Node, nodeMap ClientNodeMapper, subMap SubMapper, currentId string, options ...grpc.DialOption) {
	for n := range nodeChannel {
		cn, err := newClientNode(n, currentId, options...)
		if err != nil {
			log.Printf("skipping %s, couldn't create client node: %s", n.Id, err)
			continue
		}

		nodeMap.Add(*cn)
		subMap.Add(n.Id, n.SubscribedSubjects...)
	}
}

// listenForStaleNodes takes nodes passed in and removes them from a NodeMapper and SubMapper
func listenForStaleNodes(staleChannel chan Node, nodeMap ClientNodeMapper, subMap SubMapper) {
	for n := range staleChannel {
		nodeMap.Remove(n.Id)
		subMap.Remove(n.Id, n.SubscribedSubjects...)
	}
}

// generateIdsBySubject takes a subject string and returns a channel of node ids subscribing to it
func generateIdsBySubject(subject string, subMap SubMapper) (<-chan string, error) {
	out := make(chan string)

	ids, err := subMap.Subscribers(subject)
	if err != nil {
		return nil, fmt.Errorf("could not get subscribers: %s", err)
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
		return nil, fmt.Errorf("could not pop message response: %s", err)
	}

	out <- id
	close(out)

	return out, nil
}

// idToNodes takes a channel of ids and returns a channel of nodes based on the ids
func idToClientNodes(in <-chan string, nodeMap ClientNodeMapper) <-chan clientNode {
	out := make(chan clientNode)
	go func() {
		for id := range in {
			n, exist := nodeMap.Node(id)
			if !exist {
				log.Printf("node %s does not exist in nodemap - skipping", id)
				continue
			}

			out <- n
		}
		close(out)
	}()

	return out
}

// fanMessageAttempts takes a channel of courierClients and creates a goroutine for each to attempt a   Returns a channel that will return nodes that unsuccessfully sent a message
func fanMessageAttempts(in <-chan clientNode, ctx context.Context, metadata attemptMetadata, msg Message, send sendFunc) chan Node {
	out := make(chan Node)

	go func() {
		wg := &sync.WaitGroup{}

		for n := range in {
			wg.Add(1)
			go attemptMessage(ctx, n, metadata, msg, send, out, wg)
		}

		wg.Wait()
		close(out)
	}()

	return out
}

// attemptMessage takes a send function and attempts it until it succeeds or has reached maxAttempts.  Waits between attempts depends on the given interval
func attemptMessage(ctx context.Context, client clientNode, metadata attemptMetadata, msg Message, send sendFunc, nchan chan Node, wg *sync.WaitGroup) {
	defer wg.Done()

	attempts := 0
	for attempts < metadata.maxAttempts {
		err := send(ctx, msg, client)
		if err != nil {
			attempts++
			time.Sleep(metadata.waitInterval)
			continue
		}

		return
	}

	nchan <- client.Node
}

// forwardFailedConnections takes a channel of nodes and sends them through a stale channel as well as a failed connection channel.  Returns a bool channel that receives true when it's done
func forwardFailedConnections(in <-chan Node, fchan chan Node, schan chan Node) <-chan bool {
	out := make(chan bool)

	go func() {
		for n := range in {
			log.Printf("failed creating connection to %s:%s - will now be removing and blacklisting %s\n", n.Address, n.Port, n.Id)

			schan <- n
			fchan <- n
		}
		out <- true
		close(out)
	}()

	return out
}

type sendFunc func(context.Context, Message, clientNode) error

func sendPublishMessage(ctx context.Context, m Message, cc clientNode) error {
	if m.Type != PubMessage {
		return fmt.Errorf("message type must be of type PublishMessage")
	}

	_, err := cc.client.PublishMessage(ctx, &proto.PublishMessageRequest{
		Message: &proto.PublishMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

func sendRequestMessage(ctx context.Context, m Message, cc clientNode) error {
	if m.Type != ReqMessage {
		return fmt.Errorf("message type must be of type RequestMessage")
	}

	_, err := cc.client.RequestMessage(ctx, &proto.RequestMessageRequest{
		Message: &proto.RequestMessage{
			Id:      m.Id,
			NodeId:  cc.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

func sendResponseMessage(ctx context.Context, m Message, cc clientNode) error {
	if m.Type != RespMessage {
		return fmt.Errorf("message type must be of type ResponseMessage")
	}
	_, err := cc.client.ResponseMessage(ctx, &proto.ResponseMessageRequest{
		Message: &proto.ResponseMessage{
			Id:      m.Id,
			Subject: m.Subject,
			Content: m.Content,
		},
	})
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}
