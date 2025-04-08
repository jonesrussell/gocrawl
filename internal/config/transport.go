package config

import (
	"net/http"
	"time"

	"go.uber.org/fx"
)

// TransportModule provides the HTTP transport configuration
var TransportModule = fx.Module("transport",
	fx.Provide(
		fx.Annotate(
			NewHTTPTransport,
			fx.As(new(http.RoundTripper)),
		),
	),
)

// NewHTTPTransport creates a new HTTP transport with default settings.
func NewHTTPTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
