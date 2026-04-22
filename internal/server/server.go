package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mstgnz/self-hosted-serverless/internal/db"
	"github.com/mstgnz/self-hosted-serverless/internal/event"
	"github.com/mstgnz/self-hosted-serverless/internal/function"
)

var validFunctionName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// clientState tracks request timestamps for a single IP within the rate-limit window.
type clientState struct {
	requests []time.Time
	lastSeen time.Time
}

// rateLimiter is a per-IP sliding-window rate limiter.
type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientState
	limit   int
	window  time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*clientState),
		limit:   limit,
		window:  window,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	state, ok := rl.clients[ip]
	if !ok {
		state = &clientState{}
		rl.clients[ip] = state
	}

	// Drop requests outside the current window.
	valid := state.requests[:0]
	for _, t := range state.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	state.requests = valid
	state.lastSeen = now

	if len(state.requests) >= rl.limit {
		return false
	}

	state.requests = append(state.requests, now)
	return true
}

// cleanupLoop evicts IPs that haven't been seen in two windows to prevent unbounded growth.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-2 * rl.window)
		for ip, state := range rl.clients {
			if state.lastSeen.Before(cutoff) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// realIP extracts the client IP from the request, honouring common proxy headers.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.SplitN(ip, ",", 2)[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Server represents the serverless HTTP server
type Server struct {
	port      int
	apiKey    string
	server    *http.Server
	registry  *function.Registry
	eventBus  *event.Bus
	dbService *db.Service
	limiter   *rateLimiter
}

// NewServer creates a new serverless server
func NewServer(port int, registry *function.Registry) *Server {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Println("Warning: API_KEY not set — all endpoints are unauthenticated")
	}

	rateLimit := 100
	if v := os.Getenv("RATE_LIMIT_PER_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateLimit = n
		}
	}

	dbService, err := db.NewService(db.PostgreSQL)
	if err != nil {
		log.Printf("Warning: Failed to initialize database service: %v\n", err)
	}

	return &Server{
		port:      port,
		apiKey:    apiKey,
		registry:  registry,
		eventBus:  event.GetGlobalBus(),
		dbService: dbService,
		limiter:   newRateLimiter(rateLimit, time.Minute),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.public(s.handleHealth))
	mux.HandleFunc("/run/", s.protected(s.handleRunFunction))
	mux.HandleFunc("/functions", s.protected(s.handleListFunctions))
	mux.HandleFunc("/events", s.protected(s.handlePublishEvent))
	mux.HandleFunc("/db", s.protected(s.handleDatabaseQuery))
	mux.HandleFunc("/metrics", s.protected(s.handleGetMetrics))
	mux.HandleFunc("/metrics/", s.protected(s.handleGetFunctionMetrics))

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.dbService != nil {
		if err := s.dbService.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}

	return s.server.Shutdown(ctx)
}

// public wraps a handler with CORS and rate limiting (no auth)
func (s *Server) public(h http.HandlerFunc) http.HandlerFunc {
	return corsMiddleware(s.rateLimitMiddleware(h))
}

// protected wraps a handler with CORS, rate limiting, and API key auth
func (s *Server) protected(h http.HandlerFunc) http.HandlerFunc {
	return corsMiddleware(s.rateLimitMiddleware(s.authMiddleware(h)))
}

// corsMiddleware sets permissive CORS headers and handles preflight requests
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// rateLimitMiddleware rejects requests that exceed the per-IP rate limit.
func (s *Server) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.allow(realIP(r)) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// authMiddleware enforces API key authentication when API_KEY env var is set.
// Clients must send the key via the X-API-Key header or as a Bearer token.
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey == "" {
			next(w, r)
			return
		}

		key := r.Header.Get("X-API-Key")
		if key == "" {
			if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if key != s.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleRunFunction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/run/")
	if name == "" {
		http.Error(w, "Function name is required", http.StatusBadRequest)
		return
	}
	if !validFunctionName.MatchString(name) {
		http.Error(w, "Invalid function name", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var input map[string]any
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := s.registry.Execute(name, input)
	if err != nil {
		log.Printf("Error executing function %s: %v", name, err)
		http.Error(w, fmt.Sprintf("Error executing function: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	s.eventBus.Publish(ctx, event.Event{
		Type: "function.executed",
		Payload: map[string]any{
			"function": name,
			"input":    input,
			"result":   result,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

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

func (s *Server) handlePublishEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var evt event.Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	errors := s.eventBus.Publish(ctx, evt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "published",
		"errors": errors,
	})
}

func (s *Server) handleDatabaseQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.dbService == nil {
		http.Error(w, "Database service not available", http.StatusServiceUnavailable)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var request struct {
		Query string `json:"query"`
		Args  []any  `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Only allow read queries to limit blast radius if credentials are compromised.
	trimmed := strings.TrimSpace(strings.ToUpper(request.Query))
	if !strings.HasPrefix(trimmed, "SELECT") && !strings.HasPrefix(trimmed, "WITH") {
		http.Error(w, "Only SELECT queries are allowed via the HTTP API", http.StatusForbidden)
		return
	}

	rows, err := s.dbService.Query(request.Query, request.Args...)
	if err != nil {
		log.Printf("Error executing database query: %v", err)
		http.Error(w, fmt.Sprintf("Error executing database query: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Printf("Error getting column names: %v", err)
		http.Error(w, fmt.Sprintf("Error getting column names: %v", err), http.StatusInternalServerError)
		return
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			// Database drivers may return []byte for text columns; convert to string.
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		http.Error(w, fmt.Sprintf("Error iterating over rows: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"results": results,
	})
}

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.registry.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"metrics": metrics,
	})
}

func (s *Server) handleGetFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/metrics/")
	if name == "" {
		http.Error(w, "Function name is required", http.StatusBadRequest)
		return
	}

	metrics, exists := s.registry.GetFunctionMetrics(name)
	if !exists {
		http.Error(w, fmt.Sprintf("Function %s not found", name), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}
