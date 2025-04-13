// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"errors"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the search command dependencies
var Module = fx.Module("search",
	// Core dependencies
	config.Module,
	logger.Module,
	storage.Module,
	api.Module,
	fx.Invoke(func(cfg config.Interface, log logger.Interface) error {
		// Validate Elasticsearch configuration
		esConfig := cfg.GetElasticsearchConfig()
		if esConfig == nil {
			return errors.New("elasticsearch configuration is required")
		}

		// Log Elasticsearch configuration
		log.Debug("Elasticsearch configuration",
			"addresses", esConfig.Addresses,
			"hasAPIKey", esConfig.APIKey != "",
			"hasBasicAuth", esConfig.Username != "" && esConfig.Password != "",
			"tls.insecure_skip_verify", esConfig.TLS != nil && esConfig.TLS.InsecureSkipVerify,
			"tls.has_ca_file", esConfig.TLS != nil && esConfig.TLS.CAFile != "",
			"tls.has_cert_file", esConfig.TLS != nil && esConfig.TLS.CertFile != "",
			"tls.has_key_file", esConfig.TLS != nil && esConfig.TLS.KeyFile != "")

		return nil
	}),
)
