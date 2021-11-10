package courier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/platform-edn/courier/proto"
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
func WithHostname(hostname string) CourierOption {
	return func(c *Courier) {
		c.Hostname = hostname
	}
}

// ListensOnPort sets the port for the Courier service to serve on
func WithPort(port string) CourierOption {
	return func(c *Courier) {
		c.Port = port
	}
}

// WithClientContext sets the context for gRPC connections in the client
func WithClientContext(ctx context.Context) CourierOption {
	return func(c *Courier) {
		c.ClientContext = ctx
	}
}

// WithDialOption sets the dial options for gRPC connections in the client
func WithDialOptions(option ...grpc.DialOption) CourierOption {
	return func(c *Courier) {
		c.DialOptions = append(c.DialOptions, option...)
	}
}

// WithFailedMessageWaitInterval sets the time in between attempts to send a message
func WithFailedMessageWaitInterval(interval time.Duration) CourierOption {
	return func(c *Courier) {
		c.attemptMetadata.waitInterval = interval
	}
}

// WithMaxFailedMessageAttempts sets the max amount of attempts to send a message before blacklisting a node
func WithMaxFailedMessageAttempts(attempts int) CourierOption {
	return func(c *Courier) {
		c.attemptMetadata.maxAttempts = attempts
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
	Id                      string
	Hostname                string
	Port                    string
	SubscribedSubjects      []string
	BroadcastedSubjects     []string
	Observer                Observer
	attemptMetadata         attemptMetadata
	ClientContext           context.Context
	DialOptions             []grpc.DialOption
	observerChannel         chan []Node
	newNodeChannel          chan Node
	staleNodeChannel        chan Node
	failedConnectionChannel chan Node
	responseChannel         chan ResponseInfo
	blacklistNodes          NodeMapper
	currentNodes            NodeMapper
	clientNodes             NodeMapper
	clientSubscribers       SubMapper
	responses               ResponseMapper
	internalSubChannels     channelMapper
	server                  proto.MessageServerServer
	StartOnCreation         bool
}

// NewCourier creates a new Courier service
func NewCourier(options ...CourierOption) (*Courier, error) {
	c := &Courier{
		Id:                  uuid.NewString(),
		Port:                "8080",
		SubscribedSubjects:  []string{},
		BroadcastedSubjects: []string{},
		attemptMetadata: attemptMetadata{
			maxAttempts:  3,
			waitInterval: time.Second,
		},
		ClientContext:           context.TODO(),
		DialOptions:             []grpc.DialOption{},
		Observer:                nil,
		observerChannel:         make(chan []Node),
		newNodeChannel:          make(chan Node),
		staleNodeChannel:        make(chan Node),
		failedConnectionChannel: make(chan Node),
		responseChannel:         make(chan ResponseInfo),
		blacklistNodes:          NewNodeMap(),
		currentNodes:            NewNodeMap(),
		clientNodes:             NewNodeMap(),
		clientSubscribers:       newSubscriberMap(),
		responses:               newResponseMap(),
		internalSubChannels:     newChannelMap(),
		StartOnCreation:         true,
	}

	for _, option := range options {
		option(c)
	}

	if c.Observer == nil {
		return nil, errors.New("observer must be set")
	}

	if c.Hostname == "" {
		c.Hostname = localIp()
	}

	if len(c.DialOptions) == 0 {
		c.DialOptions = append(c.DialOptions, grpc.WithInsecure())
	}

	c.server = NewMessageServer(c.responseChannel, c.internalSubChannels)

	if c.StartOnCreation {
		err := c.Start()
		if err != nil {
			return nil, fmt.Errorf("could not start Courier service: %s", err)
		}
	}

	return c, nil
}

func (c *Courier) Start() error {
	lis, err := getListener(c.Port)
	if err != nil {
		return fmt.Errorf("could not create listener: %s", err)
	}

	go registerNodes(c.observerChannel, c.newNodeChannel, c.staleNodeChannel, c.failedConnectionChannel, c.blacklistNodes, c.currentNodes)
	go listenForResponseInfo(c.responseChannel, c.responses)
	go listenForNewNodes(c.newNodeChannel, c.clientNodes, c.clientSubscribers)
	go listenForStaleNodes(c.staleNodeChannel, c.clientNodes, c.clientSubscribers)
	go startMessageServer(c.server, lis)

	n := NewNode(c.Id, c.Hostname, c.Port, c.SubscribedSubjects, c.BroadcastedSubjects)

	err = c.Observer.AddNode(n)
	if err != nil {
		return fmt.Errorf("could not add node: %s", err)
	}

	return nil
}

func (c *Courier) Push(m Message) error {
	ids, err := generateIdsBySubject(m.Subject, c.clientSubscribers)
	if err != nil {
		return fmt.Errorf("couldn't generate ids: %s", err)
	}

	nodes := idToNodes(ids, c.clientNodes)
	clients := nodeToCourierClients(nodes, c.Id, c.DialOptions...)
	fanMessageAttempts(clients, c.ClientContext, c.attemptMetadata, m, sendPublishMessage)

	return nil
}

func (c *Courier) Request(m Message) error {
	ids, err := generateIdsByMessage(m.Id, c.responses)
	if err != nil {
		return fmt.Errorf("couldn't generate ids: %s", err)
	}

	nodes := idToNodes(ids, c.clientNodes)
	clients := nodeToCourierClients(nodes, c.Id, c.DialOptions...)
	fanMessageAttempts(clients, c.ClientContext, c.attemptMetadata, m, sendRequestMessage)

	return nil
}

func (c *Courier) Response(m Message) error {
	ids, err := generateIdsBySubject(m.Subject, c.clientSubscribers)
	if err != nil {
		return fmt.Errorf("couldn't generate ids: %s", err)
	}

	nodes := idToNodes(ids, c.clientNodes)
	clients := nodeToCourierClients(nodes, c.Id, c.DialOptions...)
	fanMessageAttempts(clients, c.ClientContext, c.attemptMetadata, m, sendResponseMessage)

	return nil
}

func (c *Courier) Subscribe(subject string) <-chan Message {
	channel := c.internalSubChannels.Add(subject)

	return channel
}
