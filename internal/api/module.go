package api

import (
	"encoding/json"
	"net/http"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// SearchRequest represents the structure of the search request
type SearchRequest struct {
	Query string `json:"query"`
	Index string `json:"index"`
	Size  int    `json:"size"`
}

// StartHTTPServer starts the HTTP server for search requests
func StartHTTPServer(log logger.Interface) (*http.Server, error) {
	log.Info("StartHTTPServer function called")
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Call the search function with the extracted parameters
		if err := executeSearch(req.Query, req.Index, req.Size, log); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Addr: ":8081", Handler: mux}

	return server, nil // Return the server instance
}

// executeSearch performs the search operation
func executeSearch(query, index string, size int, log logger.Interface) error {
	log.Info("Executing search", "query", query, "index", index, "size", size)
	return nil
}

// Module is the Fx module for the API
var Module = fx.Options(
	fx.Provide(
		StartHTTPServer, // Ensure this is correctly provided
	),
)
