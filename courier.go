package courier

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
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

func WithNodeStore(store observer.NodeStorer) CourierOption {
	return func(c *Courier) {
		c.Store = store
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

// WithClientContext sets the context for gRPC connections in the client
func WithClientContext(ctx context.Context) CourierOption {
	return func(c *Courier) {
		c.ClientOptions = append(c.ClientOptions, client.WithContext(ctx))
	}
}

// WithDialOption sets the dial options for gRPC connections in the client
func WithDialOption(option ...grpc.DialOption) CourierOption {
	return func(c *Courier) {
		c.ClientOptions = append(c.ClientOptions, client.WithDialOption(option...))
	}
}

// WithFailedMessageWaitInterval sets the time in between attempts to send a message
func WithFailedMessageWaitInterval(interval time.Duration) CourierOption {
	return func(c *Courier) {
		c.ClientOptions = append(c.ClientOptions, client.WithFailedWaitInterval(interval))
	}
}

// WithMaxFailedMessageAttempts sets the max amount of attempts to send a message before blacklisting a node
func WithMaxFailedMessageAttempts(attempts int) CourierOption {
	return func(c *Courier) {
		c.ClientOptions = append(c.ClientOptions, client.WithMaxFailedAttempts(attempts))
	}
}

// WithNodeStoreInterval sets the interval that the observer waits before attempting to refresh the current nodes in the Courier system
func WithObserverInterval(interval time.Duration) CourierOption {
	return func(c *Courier) {
		c.ObserverOptions = append(c.ObserverOptions, observer.WithObserverInterval(interval))
	}
}

// StartOnCreation tells the Courier service to start on creation or wait to be started.  Starts by default.
func StartOnCreation(tf bool) CourierOption {
	return func(c *Courier) {
		c.StartOnCreation = tf
	}
}

// Courier is a messaging and node discovery service
type Courier struct {
	Id                  string
	Address             string
	Port                string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
	Store               observer.NodeStorer
	Proxy               proxy.Proxyer
	Observer            observer.Observer
	Client              client.Clienter
	Server              proto.MessageServerServer
	ClientOptions       []client.ClientOption
	ObserverOptions     []observer.StoreObserverOption
	StartOnCreation     bool
}

// NewCourier creates a new Courier service
func NewCourier(store observer.NodeStorer, options ...CourierOption) (*Courier, error) {
	c := &Courier{
		clientOptions:   []client.ClientOption{},
		startOnCreation: true,
	}

	for _, option := range options {
		option(c)
	}

	if c.Node.Id == "" {
		// hostname, err := hostname()
		// if err != nil {
		// 	return nil, fmt.Errorf("could not set hostname for Courier: %s", err)
		// }

		// c.Node.Id = hostname
		c.Node.Id = uuid.NewString()
	}

	return c, nil
}

func (c *Courier) Start() error {
	s := server.NewMessageServer()

	co.messageProxy = proxy.NewMessageProxy(s.PushChannel())

	o, err := observer.NewStoreObserver()

	mc, err := client.NewMessageClient(c.clientOptions...)

	go startMessageServer(s, c.Port)

	err := store.AddNode(c.Node)
	if err != nil {
		return nil, fmt.Errorf("could not add Node to NodeStore: %s", err)
	}

	return nil
}

// Subscribe takes a subject and returns a channel that will receive messages that are sent on that channel
func (c *Courier) Subscribe(subject string) chan message.Message {
	return c.messageProxy.Subscribe(subject)
}

// PushChannel returns a channel that will take a message and send it to all services that are subscribed to it
func (c *Courier) PushChannel(subject string) chan message.Message {
	return c.messageClient.MessageChannel()
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

func hostname() (string, error) {
	name, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("could not access get hostname from os: %s", err)
	}

	return name, nil
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
