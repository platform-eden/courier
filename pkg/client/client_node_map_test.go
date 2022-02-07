package client_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/pkg/client"
	"github.com/platform-edn/courier/pkg/messaging"
	"github.com/platform-edn/courier/pkg/messaging/mocks"
	"github.com/platform-edn/courier/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func TestClientNodeMap_Node(t *testing.T) {
	tests := map[string]struct {
		node    registry.Node
		checkId string
		err     error
	}{
		"should retrieve node that's in the map": {
			node: registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{
				Id: "testId",
			}))[0],
			checkId: "testId",
			err:     nil,
		},
		"should return false when node doesn't exist": {
			node:    registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{}))[0],
			checkId: "testId",
			err:     &client.UnregisteredClientNodeError{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			clientNodeMap := client.NewClientNodeMap()

			clientNode, err := client.NewClientNode(test.node, uuid.NewString(), []client.ClientNodeOption{client.WithDialOptions(grpc.WithInsecure())}...)
			assert.NoError(err)

			clientNodeMap.AddClientNode(*clientNode)

			node, err := clientNodeMap.Node(test.checkId)
			if test.err != nil {
				errorType := test.err
				assert.ErrorAs(err, &errorType)
				return
			}

			assert.NoError(err)
			assert.EqualValues(node, clientNode)
		})
	}
}
func TestClientNodeMap_AddClientNode(t *testing.T) {
	tests := map[string]struct {
		nodes []registry.Node
		err   error
	}{
		"nodes are added underneath their id": {
			nodes: registry.RemovePointers(registry.CreateTestNodes(5, &registry.TestNodeOptions{})),
			err:   nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			clientNodeMap := client.NewClientNodeMap()

			for _, node := range test.nodes {
				clientNode, err := client.NewClientNode(node, uuid.NewString(), []client.ClientNodeOption{client.WithDialOptions(grpc.WithInsecure())}...)
				assert.NoError(err)

				clientNodeMap.AddClientNode(*clientNode)
			}

			for _, node := range test.nodes {
				assert.EqualValues(clientNodeMap.Nodes[node.Id].Node, node)
			}
		})
	}
}

func TestClientNodeMap_Remove(t *testing.T) {
	tests := map[string]struct {
		nodes []registry.Node
		err   error
	}{
		"all nodes are removed": {
			nodes: registry.RemovePointers(registry.CreateTestNodes(5, &registry.TestNodeOptions{})),
			err:   nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			clientNodeMap := client.NewClientNodeMap()

			for _, node := range test.nodes {
				clientNode, err := client.NewClientNode(node, uuid.NewString(), []client.ClientNodeOption{client.WithDialOptions(grpc.WithInsecure())}...)
				assert.NoError(err)

				clientNodeMap.AddClientNode(*clientNode)
			}

			for _, n := range test.nodes {
				clientNodeMap.RemoveClientNode(n.Id)
			}

			assert.Zero(len(clientNodeMap.Nodes))
		})
	}
}

func TestClientNodeMap_GenerateClientIds(t *testing.T) {
	tests := map[string]struct {
		nodes            []registry.Node
		nonExistingNodes []registry.Node
		err              error
	}{
		"all nodes passed in are passed out as clientNodes": {
			nodes:            registry.RemovePointers(registry.CreateTestNodes(5, &registry.TestNodeOptions{})),
			nonExistingNodes: []registry.Node{},
			err:              nil,
		},
		"nodes that don't exist don't get passed out": {
			nodes:            registry.RemovePointers(registry.CreateTestNodes(5, &registry.TestNodeOptions{})),
			nonExistingNodes: registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{})),
			err:              nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			clientNodeMap := client.NewClientNodeMap()
			allNodes := append(test.nodes, test.nonExistingNodes...)
			in := make(chan string, len(allNodes))

			for _, node := range test.nodes {
				clientNode, err := client.NewClientNode(node, uuid.NewString(), []client.ClientNodeOption{client.WithDialOptions(grpc.WithInsecure())}...)
				assert.NoError(err)

				clientNodeMap.AddClientNode(*clientNode)
			}

			for _, node := range allNodes {
				in <- node.Id
			}

			out := clientNodeMap.GenerateClientNodes(in)
			close(in)

			done := make(chan struct{})
			go func() {
				nodes := []registry.Node{}
				for clientNode := range out {
					nodes = append(nodes, clientNode.Node)
					assert.Contains(test.nodes, clientNode.Node)
				}

				assert.Len(nodes, len(test.nodes))
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second * 3):
				t.Fatal("did not close done channel in time")
			}
		})
	}
}

func TestClientNodeMap_FanClientNodeMessaging(t *testing.T) {
	tests := map[string]struct {
		ids       []string
		currentId string
		err       error
	}{
		"all clientNodes should be sent the message and pass": {
			ids:       []string{"test1", "test2", "test3"},
			currentId: "nodeId",
			err:       nil,
		},
		"all clientNodes should be returned through the failed node channel": {
			ids:       []string{"test1", "test2", "test3"},
			currentId: "nodeId",
			err:       errors.New("bad!"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			clientNodeMap := client.NewClientNodeMap()
			ids := make(chan string, len(test.ids))
			cli := new(mocks.MessageClient)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			msg := messaging.NewPubMessage("messageId", "testSubject", []byte("testContent"))

			cli.On("PublishMessage", ctx, mock.Anything, mock.Anything).Return(&messaging.PublishMessageResponse{}, test.err)

			for _, id := range test.ids {
				node := registry.RemovePointers(registry.CreateTestNodes(1, &registry.TestNodeOptions{
					Id: id,
				}))[0]

				clientNode, err := client.NewClientNode(node, test.currentId, client.WithInsecure())
				assert.NoError(err)

				clientNode.MessageClient = cli
				clientNodeMap.AddClientNode(*clientNode)

				ids <- id
			}

			close(ids)
			failedNodes := clientNodeMap.FanClientNodeMessaging(ctx, msg, ids)

			done := make(chan struct{})
			go func() {
				nodes := []registry.Node{}
				for node := range failedNodes {
					nodes = append(nodes, node)
				}

				cli.AssertNumberOfCalls(t, "PublishMessage", len(test.ids))
				cli.AssertExpectations(t)

				if test.err != nil {
					assert.Len(nodes, len(test.ids))
				} else {
					assert.Len(nodes, 0)
				}

				close(done)
			}()

			select {
			case <-done:
				break
			case <-time.After(time.Second * 3):
				t.Fatal("did not send messages in time")
			}

		})
	}
}
