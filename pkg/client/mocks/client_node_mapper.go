// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	context "context"

	client "github.com/platform-edn/courier/pkg/client"

	messaging "github.com/platform-edn/courier/pkg/messaging"

	mock "github.com/stretchr/testify/mock"

	registry "github.com/platform-edn/courier/pkg/registry"
)

// ClientNodeMapper is an autogenerated mock type for the ClientNodeMapper type
type ClientNodeMapper struct {
	mock.Mock
}

// AddClientNode provides a mock function with given fields: _a0
func (_m *ClientNodeMapper) AddClientNode(_a0 client.ClientNode) {
	_m.Called(_a0)
}

// FanClientNodeMessaging provides a mock function with given fields: _a0, _a1, _a2
func (_m *ClientNodeMapper) FanClientNodeMessaging(_a0 context.Context, _a1 messaging.Message, _a2 <-chan string) chan registry.Node {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 chan registry.Node
	if rf, ok := ret.Get(0).(func(context.Context, messaging.Message, <-chan string) chan registry.Node); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan registry.Node)
		}
	}

	return r0
}

// Node provides a mock function with given fields: _a0
func (_m *ClientNodeMapper) Node(_a0 string) (*client.ClientNode, error) {
	ret := _m.Called(_a0)

	var r0 *client.ClientNode
	if rf, ok := ret.Get(0).(func(string) *client.ClientNode); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.ClientNode)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveClientNode provides a mock function with given fields: _a0
func (_m *ClientNodeMapper) RemoveClientNode(_a0 string) {
	_m.Called(_a0)
}
