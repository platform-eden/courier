package courier

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/platform-edn/courier/client"
	"github.com/platform-edn/courier/message"
	"github.com/platform-edn/courier/observer"
	"github.com/platform-edn/courier/proto"
	"github.com/platform-edn/courier/proxy"
	"github.com/platform-edn/courier/server"
	"google.golang.org/grpc"
)

// CourierOption is a set of options that may be passed as parameters when creating a Courier object
type CourierOption func(c *Courier)

// Subscribes sets the subjects the Courier services will listen for
func Subscribes(subjects ...string) CourierOption {
	return func(c *Courier) {
		c.SubscribedSubjects = subjects
	}
}

// Broadcasts sets the subjects Courier services will produce on
func Broadcasts(subjects ...string) CourierOption {
	return func(c *Courier) {
		c.BroadcastedSubjects = subjects
	}
}

// ListensOnAddress sets the address for the Courier service to be found on
func ListensOnAddress(address string) CourierOption {
	return func(c *Courier) {
		c.Address = address
	}
}

// ListensOnPort sets the port for the Courier service to serve on
func ListensOnPort(port string) CourierOption {
	return func(c *Courier) {
		c.Port = port
	}
}

// WithDialOption adds an option to be passed to the MessageClient when connecting to other services
func WithDialOption(option grpc.DialOption) CourierOption {
	return func(c *Courier) {
		c.grpcOptions = append(c.grpcOptions, option)
	}
}

// WithClientMessageContext sets the context for connecting to other Courier services.  A timeout can potentially be added here
func WithClientMessageContext(ctx context.Context) CourierOption {
	return func(c *Courier) {
		c.clientContext = ctx
	}
}

// WithNodeStoreInterval sets the interval that the observer waits before attempting to refresh the current nodes in the Courier system
func WithNodeStoreInterval(interval time.Duration) CourierOption {
	return func(c *Courier) {
		c.observeInterval = interval
	}
}

// Courier is a messaging and node discovery service
type Courier struct {
	BroadcastedSubjects []string
	SubscribedSubjects  []string
	Address             string
	Port                string
	observeInterval     time.Duration
	grpcOptions         []grpc.DialOption
	clientContext       context.Context
	messageProxy        *proxy.MessageProxy
	messageClient       *client.MessageClient
}

// NewCourier creates a new Courier service
func NewCourier(store observer.NodeStorer, options ...CourierOption) *Courier {
	c := &Courier{
		Address:             localIp(),
		Port:                "8080",
		BroadcastedSubjects: []string{},
		SubscribedSubjects:  []string{},
		observeInterval:     time.Second * 3,
		grpcOptions:         []grpc.DialOption{},
		clientContext:       context.Background(),
	}

	for _, option := range options {
		option(c)
	}

	s := server.NewMessageServer()
	c.messageProxy = proxy.NewMessageProxy(s.PushChannel())
	o := observer.NewStoreObserver(store, c.observeInterval, c.BroadcastedSubjects)
	c.messageClient = client.NewMessageClient(s.ResponseChannel(), o.ListenChannel(), c.clientContext, c.Address, c.Port, c.grpcOptions)
	go startMessageServer(s, c.Port)

	return c
}

// Subscribe takes a subject and returns a channel that will receive messages that are sent on that channel
func (c *Courier) Subscribe(subject string) chan message.Message {
	return c.messageProxy.Subscribe(subject)
}

// PushChannel returns a channel that will take a message and send it to all services that are subscribed to it
func (c *Courier) PushChannel(subject string) chan message.Message {
	return c.messageClient.PushChannel()
}

// localIP returns the ip address this node is currently using
func localIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// startMessageServer starts the message server on a given port
func startMessageServer(m *server.MessageServer, port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("could not listen on port %s: %s", port, err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterMessageServerServer(grpcServer, m)

	err = grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("failed serving: %s", err)
	}

	return nil
}
