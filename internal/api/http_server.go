package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"flight-event-throttler/internal/buffer"
	"flight-event-throttler/internal/metrics"
	"flight-event-throttler/pkg/logger"
)

// Server represents the HTTP API server
type Server struct {
	logger      *logger.Logger
	metrics     *metrics.Metrics
	ringBuffer  *buffer.RingBuffer
	slidingWin  *buffer.SlidingWindowBuffer
	bufferType  string
}

// NewServer creates a new HTTP server instance
func NewServer(log *logger.Logger, m *metrics.Metrics, ringBuf *buffer.RingBuffer, slidingWin *buffer.SlidingWindowBuffer, bufferType string) *Server {
	return &Server{
		logger:     log,
		metrics:    m,
		ringBuffer: ringBuf,
		slidingWin: slidingWin,
		bufferType: bufferType,
	}
}

// SetupRoutes configures all HTTP routes
func (s *Server) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/events", s.handleEvents)
	mux.HandleFunc("/events/batch", s.handleEventsBatch)
	mux.HandleFunc("/buffer/stats", s.handleBufferStats)
}

// handleHealth returns the health status of the service
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metrics.IncrementHTTPRequests()

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    s.metrics.GetUptime().String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleMetrics returns current metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metrics.IncrementHTTPRequests()

	snapshot := s.metrics.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(snapshot)
}

// handleEvents returns all current events from the buffer
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metrics.IncrementHTTPRequests()

	var events interface{}

	if s.bufferType == "ring" && s.ringBuffer != nil {
		events = s.ringBuffer.GetAll()
	} else if s.bufferType == "sliding_window" && s.slidingWin != nil {
		events = s.slidingWin.GetAll()
	} else {
		s.logger.Error("No buffer configured")
		http.Error(w, "Buffer not available", http.StatusInternalServerError)
		s.metrics.IncrementHTTPErrors()
		return
	}

	response := map[string]interface{}{
		"events":    events,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode events response: %v", err)
		s.metrics.IncrementHTTPErrors()
	}
}

// handleEventsBatch returns a batch of events from the buffer
func (s *Server) handleEventsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metrics.IncrementHTTPRequests()

	// Parse batch size from query params, default to 100
	batchSizeStr := r.URL.Query().Get("size")
	batchSize := 100
	if batchSizeStr != "" {
		if n, err := parsePositiveInt(batchSizeStr); err == nil {
			batchSize = n
		}
	}

	var events interface{}

	if s.bufferType == "ring" && s.ringBuffer != nil {
		events = s.ringBuffer.PopBatch(batchSize)
	} else if s.bufferType == "sliding_window" && s.slidingWin != nil {
		events = s.slidingWin.PopBatch(batchSize)
	} else {
		s.logger.Error("No buffer configured")
		http.Error(w, "Buffer not available", http.StatusInternalServerError)
		s.metrics.IncrementHTTPErrors()
		return
	}

	response := map[string]interface{}{
		"events":     events,
		"batch_size": batchSize,
		"timestamp":  time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode batch response: %v", err)
		s.metrics.IncrementHTTPErrors()
	}
}

// handleBufferStats returns buffer statistics
func (s *Server) handleBufferStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.metrics.IncrementHTTPRequests()

	var stats map[string]interface{}

	if s.bufferType == "ring" && s.ringBuffer != nil {
		stats = map[string]interface{}{
			"type":     "ring",
			"count":    s.ringBuffer.Count(),
			"is_full":  s.ringBuffer.IsFull(),
			"is_empty": s.ringBuffer.IsEmpty(),
		}
	} else if s.bufferType == "sliding_window" && s.slidingWin != nil {
		stats = map[string]interface{}{
			"type":     "sliding_window",
			"count":    s.slidingWin.Count(),
			"is_empty": s.slidingWin.IsEmpty(),
		}
	} else {
		s.logger.Error("No buffer configured")
		http.Error(w, "Buffer not available", http.StatusInternalServerError)
		s.metrics.IncrementHTTPErrors()
		return
	}

	stats["timestamp"] = time.Now().Unix()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Error("Failed to encode buffer stats: %v", err)
		s.metrics.IncrementHTTPErrors()
	}
}

// parsePositiveInt parses a string to a positive integer
func parsePositiveInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("value must be positive")
	}
	return n, nil
}
