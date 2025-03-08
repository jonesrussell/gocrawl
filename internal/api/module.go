package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Constants
const (
	readHeaderTimeout = 10 * time.Second // Timeout for reading headers
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
func StartHTTPServer(log logger.Interface, searchManager SearchManager) (*http.Server, error) {
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

		// Get the total count
		total, err := searchManager.Count(r.Context(), req.Index, query)
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

	server := &http.Server{
		Handler:           mux,
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
