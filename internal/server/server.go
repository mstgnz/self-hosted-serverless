package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mstgnz/self-hosted-serverless/internal/db"
	"github.com/mstgnz/self-hosted-serverless/internal/event"
	"github.com/mstgnz/self-hosted-serverless/internal/function"
)

// Server represents the serverless HTTP server
type Server struct {
	port      int
	server    *http.Server
	registry  *function.Registry
	eventBus  *event.Bus
	dbService *db.Service
}

// NewServer creates a new serverless server
func NewServer(port int, registry *function.Registry) *Server {
	// Initialize database service (using PostgreSQL by default)
	dbService, err := db.NewService(db.PostgreSQL)
	if err != nil {
		// Log error but continue without database support
		log.Printf("Warning: Failed to initialize database service: %v\n", err)
	}

	return &Server{
		port:      port,
		registry:  registry,
		eventBus:  event.GetGlobalBus(),
		dbService: dbService,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/run/", s.handleRunFunction)
	mux.HandleFunc("/functions", s.handleListFunctions)
	mux.HandleFunc("/events", s.handlePublishEvent)
	mux.HandleFunc("/db", s.handleDatabaseQuery)

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close database connection if it exists
	if s.dbService != nil {
		if err := s.dbService.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}

	return s.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleRunFunction handles function execution requests
func (s *Server) handleRunFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract function name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/run/")
	if path == "" {
		http.Error(w, "Function name is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var input map[string]any
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute the function
	result, err := s.registry.Execute(path, input)
	if err != nil {
		log.Printf("Error executing function %s: %v", path, err)
		http.Error(w, fmt.Sprintf("Error executing function: %v", err), http.StatusInternalServerError)
		return
	}

	// Publish function execution event
	ctx := r.Context()
	s.eventBus.Publish(ctx, event.Event{
		Type: "function.executed",
		Payload: map[string]any{
			"function": path,
			"input":    input,
			"result":   result,
		},
	})

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// handleListFunctions handles listing all available functions
func (s *Server) handleListFunctions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	functions := s.registry.ListFunctions()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"functions": functions,
	})
}

// handlePublishEvent handles publishing events
func (s *Server) handlePublishEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var evt event.Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Publish the event
	ctx := r.Context()
	errors := s.eventBus.Publish(ctx, evt)

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "published",
		"errors": errors,
	})
}

// handleDatabaseQuery handles database query requests
func (s *Server) handleDatabaseQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if database service is available
	if s.dbService == nil {
		http.Error(w, "Database service not available", http.StatusServiceUnavailable)
		return
	}

	// Parse request body
	var request struct {
		Query string `json:"query"`
		Args  []any  `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute the query
	rows, err := s.dbService.Query(request.Query, request.Args...)
	if err != nil {
		log.Printf("Error executing database query: %v", err)
		http.Error(w, fmt.Sprintf("Error executing database query: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("Error getting column names: %v", err)
		http.Error(w, fmt.Sprintf("Error getting column names: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare result
	var results []map[string]any
	for rows.Next() {
		// Create a slice of any to hold the values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the values
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}

		// Create a map for this row
		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			// Handle null values
			if val == nil {
				row[col] = nil
			} else {
				// Convert to string for simplicity
				row[col] = fmt.Sprintf("%v", val)
			}
		}
		results = append(results, row)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		http.Error(w, fmt.Sprintf("Error iterating over rows: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"results": results,
	})
}
