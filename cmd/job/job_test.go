package job_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
)

// Mock implementations
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Fatal(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Printf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) GetAppConfig() *config.AppConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*config.AppConfig)
}

func (m *mockConfig) GetCommand() string {
	args := m.Called()
	if args.Get(0) == nil {
		return ""
	}
	return args.Get(0).(string)
}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*config.CrawlerConfig)
}

func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*config.ElasticsearchConfig)
}

func (m *mockConfig) GetLogConfig() *config.LogConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*config.LogConfig)
}

func (m *mockConfig) GetSources() []config.Source {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]config.Source)
}

func (m *mockConfig) GetServerConfig() *config.ServerConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*config.ServerConfig)
}

// Helper functions for creating mocks
func newMockLogger(t *testing.T) *mockLogger {
	m := &mockLogger{}
	m.Test(t)
	return m
}

func newMockConfig(t *testing.T) *mockConfig {
	m := &mockConfig{}
	m.Test(t)
	return m
}

// RunJobCommand is a helper function that runs the job command with the given dependencies
func RunJobCommand(ctx context.Context, cfg *mockConfig, logger *mockLogger, handler *signal.SignalHandler, done chan struct{}) error {
	// Create a new Fx application with the dependencies
	app := fx.New(
		fx.Supply(logger),
		fx.Supply(cfg),
		fx.Supply(done),
		fx.Supply(ctx),
		job.Module,
		fx.NopLogger,
	)

	// Start the application
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Create the job command
	cmd := job.NewJobCommand(job.JobCommandDeps{
		Logger: logger,
		Config: cfg,
	})

	// Create a new context with the signal handler
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Run the command
	err := cmd.ExecuteContext(ctx)

	// Stop the application
	if stopErr := app.Stop(ctx); stopErr != nil {
		if err != nil {
			return fmt.Errorf("failed to stop application: %w (original error: %v)", stopErr, err)
		}
		return fmt.Errorf("failed to stop application: %w", stopErr)
	}

	return err
}

type testCase struct {
	name           string
	setup          func() (*mockLogger, *mockConfig, *signal.SignalHandler)
	expectedError  error
	shouldComplete bool
	timeout        time.Duration
}

func TestJobCommand(t *testing.T) {
	tests := []testCase{
		{
			name: "successful_initialization",
			setup: func() (*mockLogger, *mockConfig, *signal.SignalHandler) {
				mockLogger := newMockLogger(t)
				mockLogger.On("Info", "Job completed").Once()
				mockLogger.On("Info", "Context cancelled, initiating shutdown...").Once()
				mockLogger.On("Info", "Shutdown requested").Once()

				mockConfig := newMockConfig(t)
				mockConfig.On("GetAppConfig").Return(&config.AppConfig{}).Once()
				mockConfig.On("GetCrawlerConfig").Return(&config.CrawlerConfig{
					BaseURL:  "https://example.com",
					MaxDepth: 2,
				}).Once()
				mockConfig.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{}).Once()
				mockConfig.On("GetLogConfig").Return(&config.LogConfig{}).Once()
				mockConfig.On("GetSources").Return([]config.Source{}).Once()
				mockConfig.On("GetServerConfig").Return(&config.ServerConfig{}).Once()
				mockConfig.On("GetCommand").Return("job").Once()

				mockHandler := signal.NewSignalHandler(mockLogger)
				mockHandler.SetShutdownTimeout(100 * time.Millisecond)

				return mockLogger, mockConfig, mockHandler
			},
			expectedError:  nil,
			shouldComplete: true,
			timeout:        2 * time.Second,
		},
		{
			name: "config_error_handling",
			setup: func() (*mockLogger, *mockConfig, *signal.SignalHandler) {
				mockLogger := newMockLogger(t)
				mockLogger.On("Info", "Context cancelled, initiating shutdown...").Once()
				mockLogger.On("Info", "Shutdown requested").Once()

				mockConfig := newMockConfig(t)
				mockConfig.On("GetAppConfig").Return(&config.AppConfig{}).Once()
				mockConfig.On("GetCrawlerConfig").Return(nil).Once()
				mockConfig.On("GetCommand").Return("job").Once()

				mockHandler := signal.NewSignalHandler(mockLogger)
				mockHandler.SetShutdownTimeout(100 * time.Millisecond)

				return mockLogger, mockConfig, mockHandler
			},
			expectedError:  errors.New("invalid crawler config"),
			shouldComplete: false,
			timeout:        2 * time.Second,
		},
		{
			name: "graceful_shutdown",
			setup: func() (*mockLogger, *mockConfig, *signal.SignalHandler) {
				mockLogger := newMockLogger(t)
				mockLogger.On("Info", "Job completed").Once()
				mockLogger.On("Info", "Context cancelled, initiating shutdown...").Once()
				mockLogger.On("Info", "Shutdown requested").Once()

				mockConfig := newMockConfig(t)
				mockConfig.On("GetAppConfig").Return(&config.AppConfig{}).Once()
				mockConfig.On("GetCrawlerConfig").Return(&config.CrawlerConfig{
					BaseURL:  "https://example.com",
					MaxDepth: 2,
				}).Once()
				mockConfig.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{}).Once()
				mockConfig.On("GetLogConfig").Return(&config.LogConfig{}).Once()
				mockConfig.On("GetSources").Return([]config.Source{}).Once()
				mockConfig.On("GetServerConfig").Return(&config.ServerConfig{}).Once()
				mockConfig.On("GetCommand").Return("job").Once()

				mockHandler := signal.NewSignalHandler(mockLogger)
				mockHandler.SetShutdownTimeout(100 * time.Millisecond)

				return mockLogger, mockConfig, mockHandler
			},
			expectedError:  nil,
			shouldComplete: true,
			timeout:        2 * time.Second,
		},
		{
			name: "shutdown_timeout",
			setup: func() (*mockLogger, *mockConfig, *signal.SignalHandler) {
				mockLogger := newMockLogger(t)
				mockLogger.On("Info", "Job completed").Once()
				mockLogger.On("Info", "Context cancelled, initiating shutdown...").Once()
				mockLogger.On("Info", "Shutdown requested").Once()

				mockConfig := newMockConfig(t)
				mockConfig.On("GetAppConfig").Return(&config.AppConfig{}).Once()
				mockConfig.On("GetCrawlerConfig").Return(&config.CrawlerConfig{
					BaseURL:  "https://example.com",
					MaxDepth: 2,
				}).Once()
				mockConfig.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{}).Once()
				mockConfig.On("GetLogConfig").Return(&config.LogConfig{}).Once()
				mockConfig.On("GetSources").Return([]config.Source{}).Once()
				mockConfig.On("GetServerConfig").Return(&config.ServerConfig{}).Once()
				mockConfig.On("GetCommand").Return("job").Once()

				mockHandler := signal.NewSignalHandler(mockLogger)
				mockHandler.SetShutdownTimeout(100 * time.Millisecond)

				return mockLogger, mockConfig, mockHandler
			},
			expectedError:  nil,
			shouldComplete: true,
			timeout:        2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, mockConfig, mockHandler := tt.setup()

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Run the job command
			err := RunJobCommand(ctx, mockConfig, mockLogger, mockHandler, nil)

			// Verify the error matches expectations
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockLogger.AssertExpectations(t)
			mockConfig.AssertExpectations(t)
		})
	}
}
