// Code generated by MockGen. DO NOT EDIT.
// Source: internal/sources/interface.go

// Package sources is a generated GoMock package.
package sources

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/jonesrussell/gocrawl/internal/config/types"
	sourceutils "github.com/jonesrussell/gocrawl/internal/sourceutils"
	types0 "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// AddSource mocks base method.
func (m *MockInterface) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddSource", ctx, source)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddSource indicates an expected call of AddSource.
func (mr *MockInterfaceMockRecorder) AddSource(ctx, source interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSource", reflect.TypeOf((*MockInterface)(nil).AddSource), ctx, source)
}

// DeleteSource mocks base method.
func (m *MockInterface) DeleteSource(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSource", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSource indicates an expected call of DeleteSource.
func (mr *MockInterfaceMockRecorder) DeleteSource(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSource", reflect.TypeOf((*MockInterface)(nil).DeleteSource), ctx, name)
}

// FindByName mocks base method.
func (m *MockInterface) FindByName(name string) *sourceutils.SourceConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByName", name)
	ret0, _ := ret[0].(*sourceutils.SourceConfig)
	return ret0
}

// FindByName indicates an expected call of FindByName.
func (mr *MockInterfaceMockRecorder) FindByName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByName", reflect.TypeOf((*MockInterface)(nil).FindByName), name)
}

// GetMetrics mocks base method.
func (m *MockInterface) GetMetrics() sourceutils.SourcesMetrics {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetrics")
	ret0, _ := ret[0].(sourceutils.SourcesMetrics)
	return ret0
}

// GetMetrics indicates an expected call of GetMetrics.
func (mr *MockInterfaceMockRecorder) GetMetrics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetrics", reflect.TypeOf((*MockInterface)(nil).GetMetrics))
}

// GetSources mocks base method.
func (m *MockInterface) GetSources() ([]sourceutils.SourceConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSources")
	ret0, _ := ret[0].([]sourceutils.SourceConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSources indicates an expected call of GetSources.
func (mr *MockInterfaceMockRecorder) GetSources() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSources", reflect.TypeOf((*MockInterface)(nil).GetSources))
}

// ListSources mocks base method.
func (m *MockInterface) ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSources", ctx)
	ret0, _ := ret[0].([]*sourceutils.SourceConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSources indicates an expected call of ListSources.
func (mr *MockInterfaceMockRecorder) ListSources(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSources", reflect.TypeOf((*MockInterface)(nil).ListSources), ctx)
}

// UpdateSource mocks base method.
func (m *MockInterface) UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSource", ctx, source)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSource indicates an expected call of UpdateSource.
func (mr *MockInterfaceMockRecorder) UpdateSource(ctx, source interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSource", reflect.TypeOf((*MockInterface)(nil).UpdateSource), ctx, source)
}

// ValidateSource mocks base method.
func (m *MockInterface) ValidateSource(ctx context.Context, sourceName string, indexManager types0.IndexManager) (*types.Source, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateSource", ctx, sourceName, indexManager)
	ret0, _ := ret[0].(*types.Source)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateSource indicates an expected call of ValidateSource.
func (mr *MockInterfaceMockRecorder) ValidateSource(ctx, sourceName, indexManager interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateSource", reflect.TypeOf((*MockInterface)(nil).ValidateSource), ctx, sourceName, indexManager)
}
