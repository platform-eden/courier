// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	messaging "github.com/platform-edn/courier/pkg/messaging"

	mock "github.com/stretchr/testify/mock"

	registry "github.com/platform-edn/courier/pkg/registry"

	sync "sync"
)

// Messager is an autogenerated mock type for the Messager type
type Messager struct {
	mock.Mock
}

// ListenForNodeEvents provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *Messager) ListenForNodeEvents(_a0 context.Context, _a1 *sync.WaitGroup, _a2 <-chan registry.NodeEvent, _a3 chan error, _a4 string) {
	_m.Called(_a0, _a1, _a2, _a3, _a4)
}

// ListenForResponseInfo provides a mock function with given fields: _a0, _a1, _a2
func (_m *Messager) ListenForResponseInfo(_a0 context.Context, _a1 *sync.WaitGroup, _a2 <-chan messaging.ResponseInfo) {
	_m.Called(_a0, _a1, _a2)
}

// Publish provides a mock function with given fields: _a0, _a1
func (_m *Messager) Publish(_a0 context.Context, _a1 messaging.Message) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, messaging.Message) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Request provides a mock function with given fields: _a0, _a1
func (_m *Messager) Request(_a0 context.Context, _a1 messaging.Message) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, messaging.Message) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Response provides a mock function with given fields: _a0, _a1
func (_m *Messager) Response(_a0 context.Context, _a1 messaging.Message) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, messaging.Message) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
