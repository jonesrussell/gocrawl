package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Constants
const (
	readHeaderTimeout = 10 * time.Second // Timeout for reading headers
	defaultPort       = "8080"           // Default port if not specified in config or env
)

// SearchRequest represents the structure of the search request
type SearchRequest struct {
	Query string `json:"query"`
	Index string `json:"index"`
	Size  int    `json:"size"`
}

// SearchResponse represents the structure of the search response
type SearchResponse struct {
	Results []interface{} `json:"results"`
	Total   int           `json:"total"`
}

// StartHTTPServer starts the HTTP server for search requests
func StartHTTPServer(log logger.Interface, searchManager SearchManager, cfg config.Interface) (*http.Server, error) {
	log.Info("StartHTTPServer function called")
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var req SearchRequest
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Build the search query
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"content": req.Query,
				},
			},
			"size": req.Size,
		}

		// Use the search manager to perform the search
		results, err := searchManager.Search(r.Context(), req.Index, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the total count using a wrapped query
		total, err := searchManager.Count(r.Context(), req.Index, map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"content": req.Query,
				},
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Prepare and send the response
		response := SearchResponse{
			Results: results,
			Total:   int(total),
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Determine server address with priority:
	// 1. GOCRAWL_PORT environment variable
	// 2. Server config from config file
	// 3. Default port
	var port string
	if envPort := os.Getenv("GOCRAWL_PORT"); envPort != "" {
		port = envPort
	} else if serverCfg := cfg.GetServerConfig(); serverCfg != nil && serverCfg.Address != "" {
		port = strings.TrimPrefix(serverCfg.Address, ":")
	} else {
		port = defaultPort
	}

	// Ensure port has colon prefix
	address := ":" + strings.TrimPrefix(port, ":")

	log.Info("Server configuration", "address", address)

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadTimeout:       cfg.GetServerConfig().ReadTimeout,
		WriteTimeout:      cfg.GetServerConfig().WriteTimeout,
		IdleTimeout:       cfg.GetServerConfig().IdleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return server, nil
}

// Module provides API dependencies
var Module = fx.Module("api",
	fx.Provide(
		StartHTTPServer,
	),
)
