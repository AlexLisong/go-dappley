// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import block "github.com/dappley/go-dappley/core/block"

import mock "github.com/stretchr/testify/mock"

// Consensus is an autogenerated mock type for the Consensus type
type Consensus struct {
	mock.Mock
}

// AddProducer provides a mock function with given fields: _a0
func (_m *Consensus) AddProducer(_a0 string) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLibProducerNum provides a mock function with given fields:
func (_m *Consensus) GetLibProducerNum() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetProducerAddress provides a mock function with given fields:
func (_m *Consensus) GetProducerAddress() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetProducers provides a mock function with given fields:
func (_m *Consensus) GetProducers() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// IsBypassingLibCheck provides a mock function with given fields:
func (_m *Consensus) IsBypassingLibCheck() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsNonRepeatingBlockProducerRequired provides a mock function with given fields:
func (_m *Consensus) IsNonRepeatingBlockProducerRequired() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsProducedLocally provides a mock function with given fields: _a0
func (_m *Consensus) IsProducedLocally(_a0 *block.Block) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*block.Block) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Validate provides a mock function with given fields: _a0
func (_m *Consensus) Validate(_a0 *block.Block) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*block.Block) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
