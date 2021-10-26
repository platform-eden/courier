package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/node"
	"github.com/platform-edn/courier/proto"
	"google.golang.org/grpc"
)

type Clienter interface {
	InfoChannel() chan node.ResponseInfo
	PushResponse(node.ResponseInfo)
	NodeChannel() chan map[string]node.Node
	UpdateNodes(map[string]node.Node)
	MessageChannel() chan message.Message
	SubjectSubscribers(string) ([]*node.Node, error)
	Sender(senderType) Sender
	Response(string) (string, error)
	Subscriber(string) (*node.Node, error)
	DialOptions() []grpc.DialOption
	GRPCContext() context.Context
	MaxFailedAttempts() int
	RemoveSubscriber(*node.Node) error
	FailedConnectionChannel() chan node.Node
	FailedWaitInterval() time.Duration
}

func ListenForResponseInfo(client Clienter) {
	for response := range client.InfoChannel() {
		client.PushResponse(response)
	}
}

func ListenForSubscribers(client Clienter) {
	for nodes := range client.NodeChannel() {
		client.UpdateNodes(nodes)
	}
}

func ListenForOutgoingMessages(client Clienter) {
	for m := range client.MessageChannel() {
		switch m.Type {
		case message.PubMessage:
			go AttemptPublishMessage(m, client)
		case message.ReqMessage:
			go AttemptRequestMessage(m, client)
		case message.RespMessage:
			go AttemptResponseMessage(m, client)
		}

	}
}

func AttemptPublishMessage(m message.Message, client Clienter) {
	nodes, err := client.SubjectSubscribers(m.Subject)
	if err != nil {
		log.Printf("could not get subject subscribers: %s", err)
	}

	sender := client.Sender(PublishSender)
	for _, n := range nodes {
		go AttemptMessage(m, n, sender, client)
	}
}

func AttemptRequestMessage(m message.Message, client Clienter) {
	nodes, err := client.SubjectSubscribers(m.Subject)
	if err != nil {
		log.Printf("could not get subject subscribers: %s", err)
	}

	sender := client.Sender(RequestSender)
	for _, n := range nodes {
		go AttemptMessage(m, n, sender, client)
	}
}

func AttemptResponseMessage(m message.Message, client Clienter) {
	id, err := client.Response(m.Id)
	if err != nil {
		log.Printf("could not get node id for message id %s: %s", m.Id, err)
	}

	n, err := client.Subscriber(id)
	if err != nil {
		log.Printf("could node with id %s: %s", id, err)
	}

	sender := client.Sender(ResponseSender)

	go AttemptMessage(m, n, sender, client)
}

// attemptMessage attempts to invoke a Sender's Send method.  If no errors are returned, this method exits.
// If errors are returned, attemptMessage will wait a designated amount of time before attempting to send again.
// This method will keep attempting to send a message until the max amount of attempts are reached, then it will remove
// the receiver node from the subscriberMap and send the node to anyone listening on the failedConnChannel
func AttemptMessage(m message.Message, receiver *node.Node, sender Sender, client Clienter) {
	attempt := 0
	for {
		c, conn, err := createGRPCClient(receiver.Address, receiver.Port, client.DialOptions()...)
		if err == nil {
			err = sender.Send(client.GRPCContext(), m, c)
			if err == nil {
				conn.Close()
				break
			}
		}

		conn.Close()
		if attempt >= client.MaxFailedAttempts() {
			log.Printf("failed creating connection to %s:%s: %s\n", receiver.Address, receiver.Port, err)
			log.Printf("will now be removing and blacklisting %s\n", receiver.Id)

			err := client.RemoveSubscriber(receiver)
			if err != nil {
				log.Print(err.Error())
			}

			client.FailedConnectionChannel() <- *receiver
			break
		}

		time.Sleep(client.FailedWaitInterval())
		attempt++
		continue
	}
}

// createGRPCClient takes an address, port, and dialOptions and returns a client connection
func createGRPCClient(address string, port string, options ...grpc.DialOption) (proto.MessageServerClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), options...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial %s: %v", fmt.Sprintf("%s:%s", address, port), err)
	}

	c := proto.NewMessageServerClient(conn)

	return c, conn, nil
}
