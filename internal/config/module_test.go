package config

import "go.uber.org/fx"

// TestModule provides a test configuration module that loads the test environment file.
// This module should only be used in tests.
var TestModule = fx.Options(
	fx.Provide(
		provideConfig(".env.test"), // Provide the config interface for tests
		NewHTTPTransport,           // Provides HTTP transport configuration
	),
)
