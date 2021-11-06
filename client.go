package courier

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type courierClient struct {
	client     proto.MessageServerClient
	connection grpc.ClientConn
	receiver   node.Node
}

type attemptMetadata struct {
	maxAttempts  int
	waitInterval time.Duration
}

// listenForResponseInfo takes ResponseInfo through a channel and pushes them into a responseMap
func listenForResponseInfo(responses chan node.ResponseInfo, respMap ResponseMapper) {
	for response := range responses {
		respMap.Push(response)
	}
}

// listenForNewNodes takes nodes passed through a channel and adds them to a NodeMapper and SubMapper
func listenForNewNodes(nodeChannel chan node.Node, nodeMap NodeMapper, subMap SubMapper) {
	for n := range nodeChannel {
		nodeMap.Add(n)
		subMap.Add(n.Id, n.SubscribedSubjects...)
	}
}

// listenForStaleNodes takes nodes passed in and removes them from a NodeMapper and SubMapper
func listenForStaleNodes(staleChannel chan node.Node, nodeMap NodeMapper, subMap SubMapper) {
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
func idToNodes(in <-chan string, nodeMap NodeMapper) <-chan node.Node {
	out := make(chan node.Node)
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

// nodeToCourierClients takes a channel of nodes and returns a channel of courierClients
func nodeToCourierClients(in <-chan node.Node, options ...grpc.DialOption) <-chan courierClient {
	out := make(chan courierClient)
	go func() {
		for n := range in {
			conn, err := grpc.Dial(fmt.Sprintf("%s:%s", n.Address, n.Port), options...)
			if err != nil {
				log.Printf("skipping node %s - could not create connection at %s: %s", n.Id, fmt.Sprintf("%s:%s", n.Address, n.Port), err)
			}

			cc := courierClient{
				client:     proto.NewMessageServerClient(conn),
				connection: *conn,
				receiver:   n,
			}

			out <- cc
		}
		close(out)
	}()

	return out
}

type sendFunc func(context.Context, message.Message, courierClient) error

// fanMessageAttempts takes a channel of courierClients and creates a goroutine for each to attempt a message.  Returns a channel that will return nodes that unsuccessfully sent a message
func fanMessageAttempts(in <-chan courierClient, ctx context.Context, metadata attemptMetadata, msg message.Message, send sendFunc) chan node.Node {
	out := make(chan node.Node)

	go func() {
		wg := &sync.WaitGroup{}

		for client := range in {
			wg.Add(1)
			go attemptMessage(ctx, client, metadata, msg, send, out, wg)
		}

		wg.Wait()
		close(out)
	}()

	return out
}

// attemptMessage takes a send function and attempts it until it succeeds or has reached maxAttempts.  Waits between attempts depends on the given interval
func attemptMessage(ctx context.Context, client courierClient, metadata attemptMetadata, msg message.Message, send sendFunc, nchan chan node.Node, wg *sync.WaitGroup) {
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

	nchan <- client.receiver
}

// forwardFailedConnections takes a channel of nodes and sends them through a stale channel as well as a failed connection channel.  Returns a bool channel that receives true when it's done
func forwardFailedConnections(in <-chan node.Node, fchan chan node.Node, schan chan node.Node) <-chan bool {
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
