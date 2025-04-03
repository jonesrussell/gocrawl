package config

import (
	"net/http"

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
