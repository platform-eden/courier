// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SubMapper is an autogenerated mock type for the SubMapper type
type SubMapper struct {
	mock.Mock
}

// AddSubscriber provides a mock function with given fields: _a0, _a1
func (_m *SubMapper) AddSubscriber(_a0 string, _a1 ...string) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// GenerateIdsBySubject provides a mock function with given fields: _a0
func (_m *SubMapper) GenerateIdsBySubject(_a0 string) (<-chan string, error) {
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

// RemoveSubscriber provides a mock function with given fields: _a0, _a1
func (_m *SubMapper) RemoveSubscriber(_a0 string, _a1 ...string) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// Subscribers provides a mock function with given fields: _a0
func (_m *SubMapper) Subscribers(_a0 string) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
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
