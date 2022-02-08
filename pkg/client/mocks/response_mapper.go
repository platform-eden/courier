// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	messaging "github.com/platform-edn/courier/pkg/messaging"
	mock "github.com/stretchr/testify/mock"
)

// ResponseMapper is an autogenerated mock type for the ResponseMapper type
type ResponseMapper struct {
	mock.Mock
}

// GenerateIdsByMessage provides a mock function with given fields: _a0
func (_m *ResponseMapper) GenerateIdsByMessage(_a0 string) (<-chan string, error) {
	ret := _m.Called(_a0)

	var r0 <-chan string
	if rf, ok := ret.Get(0).(func(string) <-chan string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan string)
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

// PushResponse provides a mock function with given fields: _a0
func (_m *ResponseMapper) PushResponse(_a0 messaging.ResponseInfo) {
	_m.Called(_a0)
}
