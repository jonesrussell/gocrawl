// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/jonesrussell/gocrawl/internal/api (interfaces: SearchManager)

// Package api is a generated GoMock package.
package api

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSearchManager is a mock of SearchManager interface.
type MockSearchManager struct {
	ctrl     *gomock.Controller
	recorder *MockSearchManagerMockRecorder
}

// MockSearchManagerMockRecorder is the mock recorder for MockSearchManager.
type MockSearchManagerMockRecorder struct {
	mock *MockSearchManager
}

// NewMockSearchManager creates a new mock instance.
func NewMockSearchManager(ctrl *gomock.Controller) *MockSearchManager {
	mock := &MockSearchManager{ctrl: ctrl}
	mock.recorder = &MockSearchManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSearchManager) EXPECT() *MockSearchManagerMockRecorder {
	return m.recorder
}

// Aggregate mocks base method.
func (m *MockSearchManager) Aggregate(arg0 context.Context, arg1 string, arg2 interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Aggregate", arg0, arg1, arg2)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Aggregate indicates an expected call of Aggregate.
func (mr *MockSearchManagerMockRecorder) Aggregate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Aggregate", reflect.TypeOf((*MockSearchManager)(nil).Aggregate), arg0, arg1, arg2)
}

// Count mocks base method.
func (m *MockSearchManager) Count(arg0 context.Context, arg1 string, arg2 interface{}) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Count", arg0, arg1, arg2)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Count indicates an expected call of Count.
func (mr *MockSearchManagerMockRecorder) Count(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*MockSearchManager)(nil).Count), arg0, arg1, arg2)
}

// Search mocks base method.
func (m *MockSearchManager) Search(arg0 context.Context, arg1 string, arg2 interface{}) ([]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", arg0, arg1, arg2)
	ret0, _ := ret[0].([]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Search indicates an expected call of Search.
func (mr *MockSearchManagerMockRecorder) Search(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Search", reflect.TypeOf((*MockSearchManager)(nil).Search), arg0, arg1, arg2)
}
