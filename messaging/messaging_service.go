package messaging

import (
	"context"

	"github.com/google/uuid"
)

type messageServer interface {
	stop()
	subscribe(string) <-chan Message
}

type messageClientHandler interface {
	stop()
	publish(context.Context, Message) error
	request(context.Context, Message) error
	response(context.Context, Message) error
}

type nodeRegister interface {
	stop()
}

// MessagingServiceOption is a set of options that may be passed as parameters when creating a MessagingService object
type MessagingServiceOption func(ms *MessagingService)

// Subscribes sets the subjects the MessagingService services will listen for
func Subscribes(subjects ...string) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.SubscribedSubjects = subjects
	}
}

// Broadcasts sets the subjects MessagingService services will produce on
func Broadcasts(subjects ...string) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.BroadcastedSubjects = subjects
	}
}

// ListensOnAddress sets the address for the MessagingService service to be found on
func WithHostname(hostname string) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.Hostname = hostname
	}
}

// ListensOnPort sets the port for the MessagingService service to serve on
func WithPort(port string) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.Port = port
	}
}

func WithClientNodeOptions(options ...ClientNodeOption) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.clientNodeOptions = append(ms.clientNodeOptions, options...)
	}
}

func WithObserverChannel(channel chan []Noder) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.observerChannel = channel
	}
}

//WithMessagingServer sets the grpc server to be used for messaging between Courier services.  Should only be used for testing
func withMessagingServer(server messageServer) MessagingServiceOption {
	return func(ms *MessagingService) {
		ms.server = server
	}
}

// MessagingService is a messaging and node discovery service
type MessagingService struct {
	Id                  string
	Hostname            string
	Port                string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
	clientNodeOptions   []ClientNodeOption
	observerChannel     chan []Noder
	server              messageServer
	client              messageClientHandler
	registry            nodeRegister
	running             bool
}

// NewMessagingService creates a new MessagingService service
func NewMessagingService(options ...MessagingServiceOption) (*MessagingService, error) {
	ms := &MessagingService{
		Id:                  uuid.NewString(),
		Port:                "8080",
		SubscribedSubjects:  []string{},
		BroadcastedSubjects: []string{},
		clientNodeOptions:   []ClientNodeOption{},
		observerChannel:     make(chan []Noder),
		server:              nil,
		running:             false,
	}

	for _, option := range options {
		option(ms)
	}
	if ms.Hostname == "" {
		ms.Hostname = localIp()
	}

	if ms.observerChannel == nil {
		return nil, &NoObserverChannelError{
			Method: "NewMessagingService",
		}
	}

	responseChannel := make(chan ResponseInfo)
	staleChannel := make(chan Node)
	newChannel := make(chan Node)
	failedChannel := make(chan Node)

	var err error
	if ms.server == nil {
		ms.server, err = newMessagingServer(&messagingServerOptions{
			port:            ms.Port,
			responseChannel: responseChannel,
			startServer:     true,
		})
		if err != nil {
			return nil, &MessagingServiceStartError{
				Method: "NewMessagingService",
				Err:    err,
			}
		}
	}

	ms.client = newMessagingClient(
		&messageClientOptions{
			failedChannel:   failedChannel,
			staleChannel:    staleChannel,
			nodeChannel:     newChannel,
			responseChannel: responseChannel,
			currentId:       ms.Id,
			clientOptions:   ms.clientNodeOptions,
			startClient:     true,
		})

	ms.registry = newNodeRegistry(&nodeRegistryOptions{
		observeChannel: ms.observerChannel,
		newChannel:     newChannel,
		failedChannel:  failedChannel,
		staleChannel:   staleChannel,
		startRegistry:  true,
	})

	ms.running = true
	return ms, nil
}

func (ms *MessagingService) Stop() {
	if !ms.running {
		return
	}

	ms.server.stop()
	ms.client.stop()
	ms.registry.stop()
	ms.running = false
}

func (ms *MessagingService) Publish(ctx context.Context, subject string, content []byte) error {
	msg := NewPubMessage(uuid.NewString(), subject, content)

	err := ms.client.publish(ctx, msg)
	if err != nil {
		return &SendMessagingServiceMessageError{
			Method: "Publish",
			Err:    err,
		}
	}

	return nil
}

func (ms *MessagingService) Request(ctx context.Context, subject string, content []byte) error {
	msg := NewReqMessage(uuid.NewString(), subject, content)

	err := ms.client.request(ctx, msg)
	if err != nil {
		return &SendMessagingServiceMessageError{
			Method: "Request",
			Err:    err,
		}
	}

	return nil
}

func (ms *MessagingService) Response(ctx context.Context, id string, subject string, content []byte) error {
	msg := NewRespMessage(id, subject, content)

	err := ms.client.response(ctx, msg)
	if err != nil {
		return &SendMessagingServiceMessageError{
			Method: "Response",
			Err:    err,
		}
	}

	return nil
}

func (ms *MessagingService) Subscribe(subject string) <-chan Message {
	channel := ms.server.subscribe(subject)

	return channel
}
