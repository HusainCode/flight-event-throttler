package processor

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"flight-event-throttler/internal/model"
)

// RateLimiter controls the rate of event processing using token bucket algorithm
type RateLimiter struct {
	limiter       *rate.Limiter
	eventsPerSec  int
	burstSize     int
	mu            sync.RWMutex
	processedCount int64
	droppedCount   int64
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(eventsPerSecond, burstSize int) *RateLimiter {
	return &RateLimiter{
		limiter:      rate.NewLimiter(rate.Limit(eventsPerSecond), burstSize),
		eventsPerSec: eventsPerSecond,
		burstSize:    burstSize,
	}
}

// Allow checks if an event can be processed based on rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	allowed := rl.limiter.Allow()
	if allowed {
		rl.processedCount++
	} else {
		rl.droppedCount++
	}
	return allowed
}

// Wait blocks until an event can be processed or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	err := rl.limiter.Wait(ctx)
	if err == nil {
		rl.mu.Lock()
		rl.processedCount++
		rl.mu.Unlock()
	}
	return err
}

// AllowN checks if n events can be processed
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	allowed := rl.limiter.AllowN(time.Now(), n)
	if allowed {
		rl.processedCount += int64(n)
	} else {
		rl.droppedCount += int64(n)
	}
	return allowed
}

// Reserve reserves a token for future use and returns a Reservation
func (rl *RateLimiter) Reserve() *rate.Reservation {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.processedCount++
	return rl.limiter.Reserve()
}

// UpdateLimit dynamically updates the rate limit
func (rl *RateLimiter) UpdateLimit(eventsPerSecond int, burstSize int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.eventsPerSec = eventsPerSecond
	rl.burstSize = burstSize
	rl.limiter.SetLimit(rate.Limit(eventsPerSecond))
	rl.limiter.SetBurst(burstSize)
}

// GetStats returns current statistics
func (rl *RateLimiter) GetStats() (processed, dropped int64) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.processedCount, rl.droppedCount
}

// Reset resets the statistics
func (rl *RateLimiter) ResetStats() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.processedCount = 0
	rl.droppedCount = 0
}

// GetLimit returns current rate limit settings
func (rl *RateLimiter) GetLimit() (eventsPerSec int, burstSize int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.eventsPerSec, rl.burstSize
}

// EventProcessor handles event processing with rate limiting and buffering
type EventProcessor struct {
	rateLimiter *RateLimiter
	inputChan   chan *model.FlightEvent
	outputChan  chan *model.FlightEvent
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(rateLimiter *RateLimiter, bufferSize int) *EventProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventProcessor{
		rateLimiter: rateLimiter,
		inputChan:   make(chan *model.FlightEvent, bufferSize),
		outputChan:  make(chan *model.FlightEvent, bufferSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins processing events
func (ep *EventProcessor) Start() {
	ep.wg.Add(1)
	go ep.processEvents()
}

// processEvents is the main processing loop
func (ep *EventProcessor) processEvents() {
	defer ep.wg.Done()

	for {
		select {
		case <-ep.ctx.Done():
			return
		case event := <-ep.inputChan:
			if event == nil {
				continue
			}

			// Wait for rate limiter to allow processing
			if err := ep.rateLimiter.Wait(ep.ctx); err != nil {
				// Context cancelled
				return
			}

			// Send to output channel
			select {
			case ep.outputChan <- event:
			case <-ep.ctx.Done():
				return
			}
		}
	}
}

// Submit submits an event for processing
func (ep *EventProcessor) Submit(event *model.FlightEvent) bool {
	select {
	case ep.inputChan <- event:
		return true
	case <-ep.ctx.Done():
		return false
	default:
		// Channel is full
		return false
	}
}

// GetOutputChannel returns the output channel for processed events
func (ep *EventProcessor) GetOutputChannel() <-chan *model.FlightEvent {
	return ep.outputChan
}

// Stop gracefully stops the processor
func (ep *EventProcessor) Stop() {
	ep.cancel()
	ep.wg.Wait()
	close(ep.inputChan)
	close(ep.outputChan)
}

// GetStats returns processing statistics
func (ep *EventProcessor) GetStats() (processed, dropped int64) {
	return ep.rateLimiter.GetStats()
}
