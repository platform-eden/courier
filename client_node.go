package courier

import (
	"fmt"

	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type clientNode struct {
	Node
	client     proto.MessageServerClient
	connection grpc.ClientConn
	currentId  string
}

func newClientNode(node Node, currrentId string, options ...grpc.DialOption) (*clientNode, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Address, node.Port), options...)
	if err != nil {
		return nil, fmt.Errorf("could not create connection at %s: %s", fmt.Sprintf("%s:%s", node.Address, node.Port), err)
	}

	n := clientNode{
		Node:       node,
		client:     proto.NewMessageServerClient(conn),
		connection: *conn,
		currentId:  currrentId,
	}

	return &n, nil
}
