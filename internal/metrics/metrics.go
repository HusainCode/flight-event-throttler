package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects and tracks system metrics
type Metrics struct {
	// Event metrics
	eventsReceived    atomic.Int64
	eventsProcessed   atomic.Int64
	eventsDropped     atomic.Int64
	eventsFailed      atomic.Int64

	// Rate metrics
	eventsPerSecond   atomic.Int64
	lastSecondCount   atomic.Int64
	lastSecondTime    atomic.Int64

	// Buffer metrics
	bufferSize        atomic.Int64
	bufferCapacity    atomic.Int64

	// API metrics
	apiRequests       atomic.Int64
	apiErrors         atomic.Int64
	apiLatencySum     atomic.Int64
	apiLatencyCount   atomic.Int64

	// HTTP metrics
	httpRequests      atomic.Int64
	httpErrors        atomic.Int64

	startTime         time.Time
	mu                sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	m := &Metrics{
		startTime: time.Now(),
	}

	// Start background ticker to calculate events per second
	go m.calculateRateMetrics()

	return m
}

// Event metrics methods

func (m *Metrics) IncrementEventsReceived() {
	m.eventsReceived.Add(1)
}

func (m *Metrics) IncrementEventsProcessed() {
	m.eventsProcessed.Add(1)
}

func (m *Metrics) IncrementEventsDropped() {
	m.eventsDropped.Add(1)
}

func (m *Metrics) IncrementEventsFailed() {
	m.eventsFailed.Add(1)
}

func (m *Metrics) GetEventsReceived() int64 {
	return m.eventsReceived.Load()
}

func (m *Metrics) GetEventsProcessed() int64 {
	return m.eventsProcessed.Load()
}

func (m *Metrics) GetEventsDropped() int64 {
	return m.eventsDropped.Load()
}

func (m *Metrics) GetEventsFailed() int64 {
	return m.eventsFailed.Load()
}

// Rate metrics methods

func (m *Metrics) GetEventsPerSecond() int64 {
	return m.eventsPerSecond.Load()
}

func (m *Metrics) calculateRateMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentProcessed := m.eventsProcessed.Load()
		lastCount := m.lastSecondCount.Load()

		rate := currentProcessed - lastCount
		m.eventsPerSecond.Store(rate)
		m.lastSecondCount.Store(currentProcessed)
	}
}

// Buffer metrics methods

func (m *Metrics) SetBufferSize(size int64) {
	m.bufferSize.Store(size)
}

func (m *Metrics) SetBufferCapacity(capacity int64) {
	m.bufferCapacity.Store(capacity)
}

func (m *Metrics) GetBufferSize() int64 {
	return m.bufferSize.Load()
}

func (m *Metrics) GetBufferCapacity() int64 {
	return m.bufferCapacity.Load()
}

func (m *Metrics) GetBufferUtilization() float64 {
	size := float64(m.bufferSize.Load())
	capacity := float64(m.bufferCapacity.Load())
	if capacity == 0 {
		return 0
	}
	return (size / capacity) * 100
}

// API metrics methods

func (m *Metrics) IncrementAPIRequests() {
	m.apiRequests.Add(1)
}

func (m *Metrics) IncrementAPIErrors() {
	m.apiErrors.Add(1)
}

func (m *Metrics) RecordAPILatency(latencyMs int64) {
	m.apiLatencySum.Add(latencyMs)
	m.apiLatencyCount.Add(1)
}

func (m *Metrics) GetAPIRequests() int64 {
	return m.apiRequests.Load()
}

func (m *Metrics) GetAPIErrors() int64 {
	return m.apiErrors.Load()
}

func (m *Metrics) GetAPIAverageLatency() float64 {
	count := m.apiLatencyCount.Load()
	if count == 0 {
		return 0
	}
	sum := m.apiLatencySum.Load()
	return float64(sum) / float64(count)
}

// HTTP metrics methods

func (m *Metrics) IncrementHTTPRequests() {
	m.httpRequests.Add(1)
}

func (m *Metrics) IncrementHTTPErrors() {
	m.httpErrors.Add(1)
}

func (m *Metrics) GetHTTPRequests() int64 {
	return m.httpRequests.Load()
}

func (m *Metrics) GetHTTPErrors() int64 {
	return m.httpErrors.Load()
}

// General metrics methods

func (m *Metrics) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.startTime)
}

func (m *Metrics) Reset() {
	m.eventsReceived.Store(0)
	m.eventsProcessed.Store(0)
	m.eventsDropped.Store(0)
	m.eventsFailed.Store(0)
	m.eventsPerSecond.Store(0)
	m.lastSecondCount.Store(0)
	m.apiRequests.Store(0)
	m.apiErrors.Store(0)
	m.apiLatencySum.Store(0)
	m.apiLatencyCount.Store(0)
	m.httpRequests.Store(0)
	m.httpErrors.Store(0)

	m.mu.Lock()
	m.startTime = time.Now()
	m.mu.Unlock()
}

// Snapshot represents a point-in-time snapshot of all metrics
type Snapshot struct {
	// Event metrics
	EventsReceived    int64   `json:"events_received"`
	EventsProcessed   int64   `json:"events_processed"`
	EventsDropped     int64   `json:"events_dropped"`
	EventsFailed      int64   `json:"events_failed"`
	EventsPerSecond   int64   `json:"events_per_second"`

	// Buffer metrics
	BufferSize        int64   `json:"buffer_size"`
	BufferCapacity    int64   `json:"buffer_capacity"`
	BufferUtilization float64 `json:"buffer_utilization_percent"`

	// API metrics
	APIRequests       int64   `json:"api_requests"`
	APIErrors         int64   `json:"api_errors"`
	APIAvgLatency     float64 `json:"api_avg_latency_ms"`

	// HTTP metrics
	HTTPRequests      int64   `json:"http_requests"`
	HTTPErrors        int64   `json:"http_errors"`

	// System metrics
	UptimeSeconds     int64   `json:"uptime_seconds"`
	Timestamp         int64   `json:"timestamp"`
}

// GetSnapshot returns a snapshot of all current metrics
func (m *Metrics) GetSnapshot() *Snapshot {
	return &Snapshot{
		EventsReceived:    m.GetEventsReceived(),
		EventsProcessed:   m.GetEventsProcessed(),
		EventsDropped:     m.GetEventsDropped(),
		EventsFailed:      m.GetEventsFailed(),
		EventsPerSecond:   m.GetEventsPerSecond(),
		BufferSize:        m.GetBufferSize(),
		BufferCapacity:    m.GetBufferCapacity(),
		BufferUtilization: m.GetBufferUtilization(),
		APIRequests:       m.GetAPIRequests(),
		APIErrors:         m.GetAPIErrors(),
		APIAvgLatency:     m.GetAPIAverageLatency(),
		HTTPRequests:      m.GetHTTPRequests(),
		HTTPErrors:        m.GetHTTPErrors(),
		UptimeSeconds:     int64(m.GetUptime().Seconds()),
		Timestamp:         time.Now().Unix(),
	}
}
