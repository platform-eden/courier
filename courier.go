package courier

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

// WithObserver sets what Observer will be looking for nodes
func WithObserver(observer Observer) CourierOption {
	return func(c *Courier) {
		c.Observer = observer
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
	DialOptions             []grpc.DialOption
	observerChannel         chan []Noder
	newNodeChannel          chan Node
	staleNodeChannel        chan Node
	failedConnectionChannel chan Node
	responseChannel         chan ResponseInfo
	waitGroup               *sync.WaitGroup
	cancelFunc              context.CancelFunc
	blacklistNodes          NodeMapper
	currentNodes            NodeMapper
	clientNodes             ClientNodeMapper
	clientSubscribers       SubMapper
	responses               ResponseMapper
	internalSubChannels     channelMapper
	server                  *grpc.Server
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
		DialOptions:             []grpc.DialOption{},
		Observer:                nil,
		observerChannel:         make(chan []Noder),
		newNodeChannel:          make(chan Node),
		staleNodeChannel:        make(chan Node),
		failedConnectionChannel: make(chan Node),
		responseChannel:         make(chan ResponseInfo),
		waitGroup:               &sync.WaitGroup{},
		blacklistNodes:          NewNodeMap(),
		currentNodes:            NewNodeMap(),
		clientNodes:             newClientNodeMap(),
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

	grpcServer := grpc.NewServer()
	proto.RegisterMessageServerServer(grpcServer, NewMessageServer(c.responseChannel, c.internalSubChannels))
	c.server = grpcServer

	if c.StartOnCreation {
		err := c.Start()
		if err != nil {
			return nil, fmt.Errorf("could not start Courier service: %s", err)
		}
	}

	return c, nil
}

func (c *Courier) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	go registerNodes(ctx, c.waitGroup, c.observerChannel, c.newNodeChannel, c.staleNodeChannel, c.failedConnectionChannel, c.blacklistNodes, c.currentNodes)
	go listenForResponseInfo(ctx, c.waitGroup, c.responseChannel, c.responses)
	go listenForNewNodes(ctx, c.waitGroup, c.newNodeChannel, c.clientNodes, c.clientSubscribers, c.Id, c.DialOptions...)
	go listenForStaleNodes(ctx, c.waitGroup, c.staleNodeChannel, c.clientNodes, c.clientSubscribers)

	err := startMessageServer(c.server, c.Port)
	if err != nil {
		return fmt.Errorf("could not start server %s", err)
	}

	n := NewNode(c.Id, c.Hostname, c.Port, c.SubscribedSubjects, c.BroadcastedSubjects)

	err = c.Observer.AddNode(n)
	if err != nil {
		return fmt.Errorf("could not add node: %s", err)
	}

	return nil
}

func (c *Courier) Stop() {
	c.server.GracefulStop()
}

func (c *Courier) Publish(ctx context.Context, subject string, content []byte) error {
	msg := NewPubMessage(uuid.NewString(), subject, content)

	ids, err := generateIdsBySubject(msg.Subject, c.clientSubscribers)
	if err != nil {
		return fmt.Errorf("couldn't generate ids: %s", err)
	}

	cnodes := idToClientNodes(ids, c.clientNodes)
	failed := fanMessageAttempts(cnodes, ctx, c.attemptMetadata, msg)
	done := forwardFailedConnections(failed, c.failedConnectionChannel, c.staleNodeChannel)

	<-done

	return nil
}

func (c *Courier) Request(ctx context.Context, subject string, content []byte) error {
	msg := NewReqMessage(uuid.NewString(), subject, content)

	ids, err := generateIdsBySubject(msg.Subject, c.clientSubscribers)
	if err != nil {
		return fmt.Errorf("couldn't generate ids: %s", err)
	}

	cnodes := idToClientNodes(ids, c.clientNodes)
	failed := fanMessageAttempts(cnodes, ctx, c.attemptMetadata, msg)
	done := forwardFailedConnections(failed, c.failedConnectionChannel, c.staleNodeChannel)

	<-done

	return nil
}

func (c *Courier) Response(ctx context.Context, id string, subject string, content []byte) error {
	msg := NewRespMessage(id, subject, content)

	ids, err := generateIdsByMessage(msg.Id, c.responses)
	if err != nil {
		return fmt.Errorf("couldn't generate ids: %s", err)
	}

	cnodes := idToClientNodes(ids, c.clientNodes)
	failed := fanMessageAttempts(cnodes, ctx, c.attemptMetadata, msg)
	done := forwardFailedConnections(failed, c.failedConnectionChannel, c.staleNodeChannel)

	<-done

	return nil
}

func (c *Courier) Subscribe(subject string) <-chan Message {
	channel := c.internalSubChannels.Add(subject)

	return channel
}
