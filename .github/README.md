# Courier
[![codecov](https://codecov.io/gh/platform-edn/courier/branch/main/graph/badge.svg?token=5IZ9HP3249)](https://codecov.io/gh/platform-edn/courier)
[![Go Report Card](https://goreportcard.com/badge/github.com/platform-eden/courier)](https://goreportcard.com/report/github.com/platform-edn/courier)
![example workflow](https://github.com/platform-edn/courier/actions/workflows/main.yaml/badge.svg)

## TODO:

- [x] create event based messaging protocol for pub/sub and request/response messages
- [ ] add kubernetes svc support for service discovery (svc messaging)
- [ ] multicontianer pod support to be accessible for other languages
- [ ] add middleware options for production readiness
- [ ] add custom resource definition support for service discovery (direct pod messaging)
- [ ] add external storage support for service discovery (database/existing service discovery methods)

## What is this?

At its core, Courier is an event based messaging service/library that prioritizes modularity and ease of use for cloud native (Kubernetes) microservices.  THe modularity of Courier allows for it to work for messaging needs across multiple use cases:

- interpod communication in a deployment
- service to service communication in a single cluster
- service to service communication across multiple clusters

With these three tiers of communication, Courier looks to handle messages for all types of distributed services.

### Why a service?

The sidecar pattern (splitting functionality of an app into multiple containers in a Kubernetes pod) has become a popular way of modularizing app components.  Courier can be deployed as a stand alone container within your application's pod so that your application handles what is being sent while Courier handles how it is being sent.  This provides you a way to create clean borders between your business logic and messaging protocol.

Another great bonus this modularization provides is the ability to impact multiple tech stacks.  By using gRPC, Courier clients are simple to create for multiple languages.  Applications written in Typescript, C#, Elixir, etc can add Courier without having to write a single line of Go.

### Why a library?

Maybe you do want to use Go!  By being built as a library first and a service second, Courier can be simple to add to a Go application.  Calling the library directly removes a step in sending and receiving your messages while maintaining an idiomatic implementation in your own code base.

## What this is not

Although the goal of Courier is to be able to give users the freedom of event based messaging, it differs from other services in the same realm like Kafka, RabbitMQ, and NATS.  The main difference between these services and Courier is that Courier doesn't have a centralized application for your services to communicate.  Instead, your applications will be using Courier to directly post messages to each other. This makes it much simpler to deploy than other messaging services, but it has the caveat of not having features like exactly once messaging, persistant logging, and consistent ordering of messages out of the box.

The current Roadmap also aims to provide a service to only those who are leveraging Kubernetes.  The way it discovers other applications is strtictly tied to resources provided by the Kubernetes environment.  With that being said, Courier is modular, and there is no reason why an Observer couldn't be added to provide functionality for other deployment strategies in the future.

## Architecture

![Courier Service Architecture](https://user-images.githubusercontent.com/51719751/147295064-0f19d075-8210-49c4-bb8c-13d7cd69609c.png)
