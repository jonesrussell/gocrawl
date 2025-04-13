// Code generated by MockGen. DO NOT EDIT.
// Source: internal/api/indexing.go

// Package api is a generated GoMock package.
package api

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDocumentManager is a mock of DocumentManager interface.
type MockDocumentManager struct {
	ctrl     *gomock.Controller
	recorder *MockDocumentManagerMockRecorder
}

// MockDocumentManagerMockRecorder is the mock recorder for MockDocumentManager.
type MockDocumentManagerMockRecorder struct {
	mock *MockDocumentManager
}

// NewMockDocumentManager creates a new mock instance.
func NewMockDocumentManager(ctrl *gomock.Controller) *MockDocumentManager {
	mock := &MockDocumentManager{ctrl: ctrl}
	mock.recorder = &MockDocumentManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDocumentManager) EXPECT() *MockDocumentManagerMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockDocumentManager) Delete(ctx context.Context, index, id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, index, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockDocumentManagerMockRecorder) Delete(ctx, index, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockDocumentManager)(nil).Delete), ctx, index, id)
}

// Get mocks base method.
func (m *MockDocumentManager) Get(ctx context.Context, index, id string) (any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, index, id)
	ret0, _ := ret[0].(any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockDocumentManagerMockRecorder) Get(ctx, index, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDocumentManager)(nil).Get), ctx, index, id)
}

// Index mocks base method.
func (m *MockDocumentManager) Index(ctx context.Context, index, id string, doc any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Index", ctx, index, id, doc)
	ret0, _ := ret[0].(error)
	return ret0
}

// Index indicates an expected call of Index.
func (mr *MockDocumentManagerMockRecorder) Index(ctx, index, id, doc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Index", reflect.TypeOf((*MockDocumentManager)(nil).Index), ctx, index, id, doc)
}

// Update mocks base method.
func (m *MockDocumentManager) Update(ctx context.Context, index, id string, doc any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, index, id, doc)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockDocumentManagerMockRecorder) Update(ctx, index, id, doc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDocumentManager)(nil).Update), ctx, index, id, doc)
}
