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

func listenForResponseInfo(responses chan node.ResponseInfo, respMap ResponseMapper) {
	for response := range responses {
		respMap.Push(response)
	}
}

func listenForNewNodes(nodeChannel chan node.Node, nodeMap NodeMapper, subMap SubMapper) {
	for n := range nodeChannel {
		nodeMap.Add(n)
		subMap.Add(n.Id, n.SubscribedSubjects...)
	}
}

func listenForStaleNodes(staleChannel chan node.Node, nodeMap NodeMapper, subMap SubMapper) {
	for n := range staleChannel {
		nodeMap.Remove(n.Id)
		subMap.Remove(n.Id, n.SubscribedSubjects...)
	}
}

func getIdsBySubject(subject string, subMap SubMapper) (<-chan string, error) {
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

func getIdsByMessage(messageId string, respMap ResponseMapper) (<-chan string, error) {
	out := make(chan string)

	id, err := respMap.Pop(messageId)
	if err != nil {
		return nil, fmt.Errorf("could not pop message response: %s", err)
	}

	go func() {
		out <- id
		close(out)
	}()

	return out, nil
}

func generateNodesById(in <-chan string, nodeMap NodeMapper) <-chan node.Node {
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

type courierClient struct {
	client     proto.MessageServerClient
	connection grpc.ClientConn
	receiver   node.Node
}

func generateCourierClient(in <-chan node.Node, options ...grpc.DialOption) <-chan courierClient {
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

type attemptData struct {
	message      message.Message
	maxAttempts  int
	waitInterval time.Duration
	ctx          context.Context
}

type sendFunc func(context.Context, message.Message, courierClient) error

func fanMessageAttempts(in <-chan courierClient, ctx context.Context, metadata attemptData, msg message.Message, send sendFunc) chan node.Node {
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

func attemptMessage(ctx context.Context, client courierClient, metadata attemptData, msg message.Message, send sendFunc, nchan chan node.Node, wg *sync.WaitGroup) {
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

func forwardFailedConnections(in <-chan node.Node, fchan chan node.Node, schan chan node.Node) <-chan bool {
	out := make(chan bool)

	for n := range in {
		log.Printf("failed creating connection to %s:%s - will now be removing and blacklisting %s\n", n.Address, n.Port, n.Id)

		schan <- n
		fchan <- n
	}

	return out
}
