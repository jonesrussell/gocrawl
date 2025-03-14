package httpd_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockConfig implements config.Interface for testing
type mockConfig struct {
	ctrl     *gomock.Controller
	recorder *mockConfigMockRecorder
}

type mockConfigMockRecorder struct {
	mock *mockConfig
}

func NewMockConfig(ctrl *gomock.Controller) *mockConfig {
	mock := &mockConfig{ctrl: ctrl}
	mock.recorder = &mockConfigMockRecorder{mock}
	return mock
}

func (m *mockConfig) EXPECT() *mockConfigMockRecorder {
	return m.recorder
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServerConfig")
	ret0, _ := ret[0].(*config.ServerConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetServerConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetServerConfig", nil)
}

func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetElasticsearchConfig")
	ret0, _ := ret[0].(*config.ElasticsearchConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetElasticsearchConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetElasticsearchConfig", nil)
}

func (m *mockConfig) GetAppConfig() *config.AppConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppConfig")
	ret0, _ := ret[0].(*config.AppConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetAppConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetAppConfig", nil)
}

func (m *mockConfig) GetLogConfig() *config.LogConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogConfig")
	ret0, _ := ret[0].(*config.LogConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetLogConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetLogConfig", nil)
}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCrawlerConfig")
	ret0, _ := ret[0].(*config.CrawlerConfig)
	return ret0
}

func (m *mockConfigMockRecorder) GetCrawlerConfig() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetCrawlerConfig", nil)
}

func (m *mockConfig) GetSources() []config.Source {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSources")
	ret0, _ := ret[0].([]config.Source)
	return ret0
}

func (m *mockConfigMockRecorder) GetSources() *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "GetSources", nil)
}

// mockSearchManager implements api.SearchManager for testing
type mockSearchManager struct {
	ctrl     *gomock.Controller
	recorder *mockSearchManagerMockRecorder
}

type mockSearchManagerMockRecorder struct {
	mock *mockSearchManager
}

func NewMockSearchManager(ctrl *gomock.Controller) *mockSearchManager {
	mock := &mockSearchManager{ctrl: ctrl}
	mock.recorder = &mockSearchManagerMockRecorder{mock}
	return mock
}

func (m *mockSearchManager) EXPECT() *mockSearchManagerMockRecorder {
	return m.recorder
}

func (m *mockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Search", ctx, index, query)
	ret0, _ := ret[0].([]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (m *mockSearchManagerMockRecorder) Search(ctx, index, query interface{}) *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "Search", nil, ctx, index, query)
}

func (m *mockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Count", ctx, index, query)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (m *mockSearchManagerMockRecorder) Count(ctx, index, query interface{}) *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "Count", nil, ctx, index, query)
}

func (m *mockSearchManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Aggregate", ctx, index, aggs)
	ret1, _ := ret[1].(error)
	return ret[0], ret1
}

func (m *mockSearchManagerMockRecorder) Aggregate(ctx, index, aggs interface{}) *gomock.Call {
	m.mock.ctrl.T.Helper()
	return m.mock.ctrl.RecordCallWithMethodType(m.mock, "Aggregate", nil, ctx, index, aggs)
}

func startServer(server *http.Server) error {
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return nil
}

func TestModule(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	// Create mock dependencies
	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
	mockLogger.EXPECT().Info("Server configuration", "address", ":8080").Times(1)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	mockCfg := NewMockConfig(ctrl)
	mockCfg.EXPECT().GetServerConfig().Return(&config.ServerConfig{
		Address:      ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}).MinTimes(1)

	mockSearch := NewMockSearchManager(ctrl)

	var server *http.Server

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() api.SearchManager { return mockSearch },
			func() config.Interface { return mockCfg },
		),
		httpd.Module,
		fx.Invoke(func(s *http.Server) error {
			server = s
			return startServer(s)
		}),
	)

	require.NoError(t, app.Err())
	app.RequireStart()
	t.Cleanup(func() {
		if server != nil {
			server.Close()
		}
		app.RequireStop()
	})

	require.NotNil(t, server)
	require.Equal(t, ":8080", server.Addr)
}

// TestModuleProvides tests that the httpd module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Info("StartHTTPServer function called").Times(1)
	mockLogger.EXPECT().Info("Server configuration", "address", ":8080").Times(1)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	serverConfig := &config.ServerConfig{
		Address:      ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mockCfg := NewMockConfig(ctrl)
	mockCfg.EXPECT().GetServerConfig().Return(serverConfig).AnyTimes()

	mockSearch := NewMockSearchManager(ctrl)

	var server *http.Server

	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface { return mockLogger },
			func() api.SearchManager { return mockSearch },
			func() config.Interface { return mockCfg },
		),
		httpd.Module,
		fx.Invoke(func(s *http.Server) {
			server = s
		}),
	)

	require.NoError(t, app.Err())
	app.Start(context.Background())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.Stop(ctx)
	})

	require.NotNil(t, server)
	require.Equal(t, ":8080", server.Addr)
}
