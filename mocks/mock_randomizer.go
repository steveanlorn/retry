// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/steveanlorn/retry (interfaces: Randomizer)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRandomizer is a mock of Randomizer interface
type MockRandomizer struct {
	ctrl     *gomock.Controller
	recorder *MockRandomizerMockRecorder
}

// MockRandomizerMockRecorder is the mock recorder for MockRandomizer
type MockRandomizerMockRecorder struct {
	mock *MockRandomizer
}

// NewMockRandomizer creates a new mock instance
func NewMockRandomizer(ctrl *gomock.Controller) *MockRandomizer {
	mock := &MockRandomizer{ctrl: ctrl}
	mock.recorder = &MockRandomizerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRandomizer) EXPECT() *MockRandomizerMockRecorder {
	return m.recorder
}

// Int63n mocks base method
func (m *MockRandomizer) Int63n(arg0 int64) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Int63n", arg0)
	ret0, _ := ret[0].(int64)
	return ret0
}

// Int63n indicates an expected call of Int63n
func (mr *MockRandomizerMockRecorder) Int63n(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Int63n", reflect.TypeOf((*MockRandomizer)(nil).Int63n), arg0)
}
