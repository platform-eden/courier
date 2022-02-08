package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/platform-edn/courier/pkg/lock"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/registry"
)

type clientNodeMap struct {
	Nodes map[string]ClientNode
	lock.Locker
}

func NewClientNodeMap() *clientNodeMap {
	nm := clientNodeMap{
		Nodes:  map[string]ClientNode{},
		Locker: lock.NewTicketLock(),
	}

	return &nm
}

func (nm *clientNodeMap) Node(id string) (*ClientNode, error) {
	nm.Lock()
	defer nm.Unlock()

	node, exist := nm.Nodes[id]
	if !exist {
		return nil, fmt.Errorf("Node: %w", &UnregisteredClientNodeError{
			Id: id,
		})
	}

	return &node, nil
}

func (nm *clientNodeMap) AddClientNode(n ClientNode) {
	nm.Lock()
	defer nm.Unlock()

	nm.Nodes[n.Id] = n
}

func (nm *clientNodeMap) RemoveClientNode(id string) {
	nm.Lock()
	defer nm.Unlock()

	delete(nm.Nodes, id)
}

func (nm *clientNodeMap) GenerateClientNodes(in <-chan string) <-chan *ClientNode {
	out := make(chan *ClientNode)
	go func() {
		for id := range in {
			node, err := nm.Node(id)
			if err != nil {
				log.Printf("%s\n", fmt.Errorf("GenerateClientNodes: %w", err))
				continue
			}

			out <- node
		}
		close(out)
	}()

	return out
}

func (nm *clientNodeMap) FanClientNodeMessaging(ctx context.Context, msg messaging.Message, ids <-chan string) chan registry.Node {
	failedNodes := make(chan registry.Node)
	clientNodes := nm.GenerateClientNodes(ids)

	go func() {
		wg := &sync.WaitGroup{}

		for client := range clientNodes {
			wg.Add(1)
			go func(client *ClientNode) {
				defer wg.Done()

				err := client.AttemptMessage(ctx, msg)
				if err != nil {
					log.Printf("%s\n", fmt.Errorf("FanClientNodeMessaging: %w", err))
					failedNodes <- client.Node
				}
			}(client)
		}

		wg.Wait()
		close(failedNodes)
	}()

	return failedNodes
}
