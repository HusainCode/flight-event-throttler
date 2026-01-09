package buffer

import (
	"sync"
	"time"

	"flight-event-throttler/internal/model"
)

// SlidingWindowBuffer stores events with timestamps and automatically removes old events
type SlidingWindowBuffer struct {
	events        []*timestampedEvent
	windowSize    time.Duration
	maxSize       int
	mu            sync.RWMutex
}

type timestampedEvent struct {
	event     *model.FlightEvent
	timestamp time.Time
}

// NewSlidingWindowBuffer creates a new sliding window buffer
func NewSlidingWindowBuffer(windowSize time.Duration, maxSize int) *SlidingWindowBuffer {
	return &SlidingWindowBuffer{
		events:     make([]*timestampedEvent, 0, maxSize),
		windowSize: windowSize,
		maxSize:    maxSize,
	}
}

// Push adds a new event to the buffer with current timestamp
func (swb *SlidingWindowBuffer) Push(event *model.FlightEvent) {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	// Remove expired events
	swb.removeExpired()

	// Add new event
	te := &timestampedEvent{
		event:     event,
		timestamp: time.Now(),
	}

	swb.events = append(swb.events, te)

	// If we exceed max size, remove oldest events
	if len(swb.events) > swb.maxSize {
		swb.events = swb.events[len(swb.events)-swb.maxSize:]
	}
}

// removeExpired removes events outside the time window (must be called with lock held)
func (swb *SlidingWindowBuffer) removeExpired() {
	if len(swb.events) == 0 {
		return
	}

	cutoffTime := time.Now().Add(-swb.windowSize)

	// Find the first non-expired event
	firstValid := 0
	for i, te := range swb.events {
		if te.timestamp.After(cutoffTime) {
			firstValid = i
			break
		}
		firstValid = i + 1
	}

	// Remove expired events
	if firstValid > 0 {
		swb.events = swb.events[firstValid:]
	}
}

// GetAll returns all events within the time window
func (swb *SlidingWindowBuffer) GetAll() []*model.FlightEvent {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	swb.removeExpired()

	events := make([]*model.FlightEvent, 0, len(swb.events))
	for _, te := range swb.events {
		events = append(events, te.event)
	}

	return events
}

// GetBatch returns up to n events from the buffer
func (swb *SlidingWindowBuffer) GetBatch(n int) []*model.FlightEvent {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	swb.removeExpired()

	count := len(swb.events)
	if n > count {
		n = count
	}

	events := make([]*model.FlightEvent, 0, n)
	for i := 0; i < n; i++ {
		events = append(events, swb.events[i].event)
	}

	return events
}

// PopBatch removes and returns up to n oldest events
func (swb *SlidingWindowBuffer) PopBatch(n int) []*model.FlightEvent {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	swb.removeExpired()

	count := len(swb.events)
	if n > count {
		n = count
	}

	if n == 0 {
		return nil
	}

	events := make([]*model.FlightEvent, 0, n)
	for i := 0; i < n; i++ {
		events = append(events, swb.events[i].event)
	}

	// Remove the popped events
	swb.events = swb.events[n:]

	return events
}

// Count returns the number of events in the window
func (swb *SlidingWindowBuffer) Count() int {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	swb.removeExpired()
	return len(swb.events)
}

// CountInLastDuration returns the number of events in a specific duration
func (swb *SlidingWindowBuffer) CountInLastDuration(duration time.Duration) int {
	swb.mu.RLock()
	defer swb.mu.RUnlock()

	cutoffTime := time.Now().Add(-duration)
	count := 0

	for i := len(swb.events) - 1; i >= 0; i-- {
		if swb.events[i].timestamp.After(cutoffTime) {
			count++
		} else {
			break
		}
	}

	return count
}

// IsEmpty returns true if the buffer is empty
func (swb *SlidingWindowBuffer) IsEmpty() bool {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	swb.removeExpired()
	return len(swb.events) == 0
}

// Clear removes all events from the buffer
func (swb *SlidingWindowBuffer) Clear() {
	swb.mu.Lock()
	defer swb.mu.Unlock()

	swb.events = make([]*timestampedEvent, 0, swb.maxSize)
}

// GetEventsInRange returns events within a specific time range
func (swb *SlidingWindowBuffer) GetEventsInRange(start, end time.Time) []*model.FlightEvent {
	swb.mu.RLock()
	defer swb.mu.RUnlock()

	events := make([]*model.FlightEvent, 0)
	for _, te := range swb.events {
		if te.timestamp.After(start) && te.timestamp.Before(end) {
			events = append(events, te.event)
		}
	}

	return events
}
