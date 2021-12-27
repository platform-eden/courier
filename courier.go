package courier

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type channelMapper interface {
	Add(string) <-chan Message
	Subscriptions(string) ([]chan Message, error)
	Close()
}

type ResponseMapper interface {
	Push(ResponseInfo)
	Pop(string) (string, error)
}

type SubMapper interface {
	Add(string, ...string)
	Remove(string, ...string)
	Subscribers(string) ([]string, error)
}

type NodeMapper interface {
	Node(string) (Node, bool)
	Nodes() map[string]Node
	Update(...Node)
	Add(Node)
	Remove(string)
	Length() int
}

type ClientNodeMapper interface {
	Node(string) (clientNode, bool)
	Add(clientNode)
	Remove(string)
	Length() int
}

type CourierServer interface {
	Start(context.Context, *sync.WaitGroup) error
}

type NoObserverError struct {
	Method string
}

func (err *NoObserverError) Error() string {
	return fmt.Sprintf("%s: observer must be set", err.Method)
}

type SendCourierMessageError struct {
	Method string
	Err    error
}

func (err *SendCourierMessageError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

type CourierStartError struct {
	Method string
	Err    error
}

func (err *CourierStartError) Error() string {
	return fmt.Sprintf("%s: %s", err.Method, err.Err)
}

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

func WithClientNodeOptions(options ...ClientNodeOption) CourierOption {
	return func(c *Courier) {
		c.clientNodeOptions = append(c.clientNodeOptions, options...)
	}
}

// WithMaxFailedMessageAttempts sets the max amount of attempts to send a message before blacklisting a node

//WithGRPCServer sets the grpc server to be used for messaging between courier services.  Should only be used for testing
func withCourierServer(server CourierServer) CourierOption {
	return func(c *Courier) {
		c.server = server
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
	clientNodeOptions       []ClientNodeOption
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
	server                  CourierServer
	StartOnCreation         bool
	running                 bool
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
		clientNodeOptions:       []ClientNodeOption{},
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
		server:                  nil,
		running:                 false,
	}

	for _, option := range options {
		option(c)
	}
	if c.Observer == nil {
		return nil, &NoObserverError{
			Method: "NewCourier",
		}
	}
	if c.Hostname == "" {
		c.Hostname = localIp()
	}
	if len(c.clientNodeOptions) == 0 {
		c.clientNodeOptions = append(c.clientNodeOptions, WithInsecure())
	}
	if c.server == nil {
		c.server = NewMessageServer(c.Port, c.responseChannel, c.internalSubChannels)
		// for canceling the server
		c.waitGroup.Add(1)
	}

	// one for each long running goroutine
	c.waitGroup.Add(4)

	if c.StartOnCreation {
		err := c.Start()
		if err != nil {
			return nil, &CourierStartError{
				Method: "NewCourier",
				Err:    err,
			}
		}
	}

	return c, nil
}

func (c *Courier) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	go registerNodes(ctx, c.waitGroup, c.observerChannel, c.newNodeChannel, c.staleNodeChannel, c.failedConnectionChannel, c.blacklistNodes, c.currentNodes)
	go listenForResponseInfo(ctx, c.waitGroup, c.responseChannel, c.responses)
	go listenForNewNodes(ctx, c.waitGroup, c.newNodeChannel, c.clientNodes, c.clientSubscribers, c.Id, c.clientNodeOptions...)
	go listenForStaleNodes(ctx, c.waitGroup, c.staleNodeChannel, c.clientNodes, c.clientSubscribers)

	err := c.server.Start(ctx, c.waitGroup)
	if err != nil {
		return &CourierStartError{
			Method: "Start",
			Err:    err,
		}
	}

	n := NewNode(c.Id, c.Hostname, c.Port, c.SubscribedSubjects, c.BroadcastedSubjects)

	err = c.Observer.AddNode(n)
	if err != nil {
		return &CourierStartError{
			Method: "Start",
			Err:    err,
		}
	}

	c.running = true
	return nil
}

func (c *Courier) Stop() {
	if !c.running {
		return
	}

	c.cancelFunc()
	c.waitGroup.Wait()
	close(c.observerChannel)
	close(c.responseChannel)
	close(c.staleNodeChannel)
	close(c.newNodeChannel)
	close(c.failedConnectionChannel)
	c.running = false
}

func (c *Courier) Publish(ctx context.Context, subject string, content []byte) error {
	msg := NewPubMessage(uuid.NewString(), subject, content)

	ids, err := generateIdsBySubject(msg.Subject, c.clientSubscribers)
	if err != nil {
		return &SendCourierMessageError{
			Method: "Publish",
			Err:    err,
		}
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
		return &SendCourierMessageError{
			Method: "Request",
			Err:    err,
		}
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
		return &SendCourierMessageError{
			Method: "Response",
			Err:    err,
		}
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
