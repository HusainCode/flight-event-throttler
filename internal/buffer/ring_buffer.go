package buffer

import (
	"sync"

	"flight-event-throttler/internal/model"
)

// RingBuffer is a circular buffer for storing flight events
type RingBuffer struct {
	buffer   []*model.FlightEvent
	size     int
	head     int
	tail     int
	count    int
	mu       sync.RWMutex
	isFull   bool
}

// NewRingBuffer creates a new ring buffer with the specified size
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buffer: make([]*model.FlightEvent, size),
		size:   size,
		head:   0,
		tail:   0,
		count:  0,
		isFull: false,
	}
}

// Push adds a new event to the buffer
// If the buffer is full, it overwrites the oldest event
func (rb *RingBuffer) Push(event *model.FlightEvent) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.head] = event
	rb.head = (rb.head + 1) % rb.size

	if rb.isFull {
		rb.tail = (rb.tail + 1) % rb.size
	} else {
		rb.count++
		if rb.head == rb.tail {
			rb.isFull = true
		}
	}
}

// Pop removes and returns the oldest event from the buffer
func (rb *RingBuffer) Pop() *model.FlightEvent {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 && !rb.isFull {
		return nil
	}

	event := rb.buffer[rb.tail]
	rb.buffer[rb.tail] = nil
	rb.tail = (rb.tail + 1) % rb.size

	if rb.isFull {
		rb.isFull = false
	} else {
		rb.count--
	}

	return event
}

// PopBatch removes and returns up to n events from the buffer
func (rb *RingBuffer) PopBatch(n int) []*model.FlightEvent {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	availableCount := rb.Count()
	if n > availableCount {
		n = availableCount
	}

	if n == 0 {
		return nil
	}

	events := make([]*model.FlightEvent, 0, n)
	for i := 0; i < n; i++ {
		if rb.count == 0 && !rb.isFull {
			break
		}

		event := rb.buffer[rb.tail]
		rb.buffer[rb.tail] = nil
		rb.tail = (rb.tail + 1) % rb.size

		if rb.isFull {
			rb.isFull = false
		} else {
			rb.count--
		}

		events = append(events, event)
	}

	return events
}

// Peek returns the oldest event without removing it
func (rb *RingBuffer) Peek() *model.FlightEvent {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 && !rb.isFull {
		return nil
	}

	return rb.buffer[rb.tail]
}

// Count returns the number of events currently in the buffer
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.isFull {
		return rb.size
	}
	return rb.count
}

// IsFull returns true if the buffer is full
func (rb *RingBuffer) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	return rb.isFull
}

// IsEmpty returns true if the buffer is empty
func (rb *RingBuffer) IsEmpty() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	return rb.count == 0 && !rb.isFull
}

// Clear removes all events from the buffer
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer = make([]*model.FlightEvent, rb.size)
	rb.head = 0
	rb.tail = 0
	rb.count = 0
	rb.isFull = false
}

// GetAll returns all events in the buffer without removing them
func (rb *RingBuffer) GetAll() []*model.FlightEvent {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.IsEmpty() {
		return nil
	}

	count := rb.Count()
	events := make([]*model.FlightEvent, 0, count)

	if rb.isFull {
		// Buffer is full, read from tail to end, then from start to head
		for i := rb.tail; i < rb.size; i++ {
			events = append(events, rb.buffer[i])
		}
		for i := 0; i < rb.head; i++ {
			events = append(events, rb.buffer[i])
		}
	} else {
		// Buffer is not full, read from tail to head
		if rb.head > rb.tail {
			for i := rb.tail; i < rb.head; i++ {
				events = append(events, rb.buffer[i])
			}
		} else {
			for i := rb.tail; i < rb.size; i++ {
				events = append(events, rb.buffer[i])
			}
			for i := 0; i < rb.head; i++ {
				events = append(events, rb.buffer[i])
			}
		}
	}

	return events
}
