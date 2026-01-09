package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flight-event-throttler/internal/api"
	"flight-event-throttler/internal/buffer"
	"flight-event-throttler/internal/config"
	"flight-event-throttler/internal/fetcher"
	"flight-event-throttler/internal/metrics"
	"flight-event-throttler/internal/model"
	"flight-event-throttler/internal/processor"
	"flight-event-throttler/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.Logging.Level)
	log.Info("Starting Flight Event Throttler...")
	log.Info("Configuration loaded successfully")

	// Initialize metrics collector
	metricsCollector := metrics.NewMetrics()
	log.Info("Metrics collector initialized")

	// Initialize buffer based on configuration
	var ringBuf *buffer.RingBuffer
	var slidingWin *buffer.SlidingWindowBuffer

	if cfg.Buffer.Type == "ring" {
		ringBuf = buffer.NewRingBuffer(cfg.Buffer.Size)
		log.Info("Ring buffer initialized with size %d", cfg.Buffer.Size)
	} else {
		slidingWin = buffer.NewSlidingWindowBuffer(cfg.RateLimit.WindowDuration, cfg.Buffer.Size)
		log.Info("Sliding window buffer initialized with size %d and window %v", cfg.Buffer.Size, cfg.RateLimit.WindowDuration)
	}

	// Update buffer metrics
	metricsCollector.SetBufferCapacity(int64(cfg.Buffer.Size))

	// Initialize rate limiter
	rateLimiter := processor.NewRateLimiter(cfg.RateLimit.EventsPerSecond, cfg.RateLimit.BurstSize)
	log.Info("Rate limiter initialized: %d events/sec, burst size %d", cfg.RateLimit.EventsPerSecond, cfg.RateLimit.BurstSize)

	// Initialize event processor
	eventProcessor := processor.NewEventProcessor(rateLimiter, cfg.Buffer.Size)
	eventProcessor.Start()
	log.Info("Event processor started")

	// Initialize OpenSky API client
	openSkyClient := fetcher.NewOpenSkyClient(
		cfg.OpenSky.BaseURL,
		cfg.OpenSky.RequestTimeout,
		cfg.OpenSky.Username,
		cfg.OpenSky.Password,
		log,
		metricsCollector,
	)
	log.Info("OpenSky API client initialized: %s", cfg.OpenSky.BaseURL)

	// Create context for managing goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start OpenSky polling in background
	go func() {
		openSkyClient.PollContinuously(ctx, cfg.OpenSky.PollInterval, func(events []*model.FlightEvent) {
			log.Debug("Received %d flight events from OpenSky API", len(events))
			metricsCollector.IncrementEventsReceived()

			for _, event := range events {
				// Add timestamp to event
				event.Timestamp = time.Now()

				// Try to submit event to processor
				if eventProcessor.Submit(event) {
					// Successfully queued
					metricsCollector.IncrementEventsProcessed()

					// Add to buffer
					if cfg.Buffer.Type == "ring" && ringBuf != nil {
						ringBuf.Push(event)
						metricsCollector.SetBufferSize(int64(ringBuf.Count()))
					} else if cfg.Buffer.Type == "sliding_window" && slidingWin != nil {
						slidingWin.Push(event)
						metricsCollector.SetBufferSize(int64(slidingWin.Count()))
					}
				} else {
					// Queue full, event dropped
					metricsCollector.IncrementEventsDropped()
					log.Debug("Event dropped: queue full")
				}
			}
		})
	}()

	// Initialize HTTP API server
	apiServer := api.NewServer(log, metricsCollector, ringBuf, slidingWin, cfg.Buffer.Type)

	// Setup HTTP routes
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Info("HTTP server starting on port %d", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	log.Info("Flight Event Throttler is running")
	log.Info("Available endpoints:")
	log.Info("  - GET /health       - Health check")
	log.Info("  - GET /metrics      - System metrics")
	log.Info("  - GET /events       - Get all buffered events")
	log.Info("  - GET /events/batch - Get batch of events")
	log.Info("  - GET /buffer/stats - Buffer statistics")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down gracefully...")

	// Cancel context to stop background goroutines
	cancel()

	// Stop event processor
	eventProcessor.Stop()
	log.Info("Event processor stopped")

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server forced to shutdown: %v", err)
	}

	// Print final metrics
	snapshot := metricsCollector.GetSnapshot()
	log.Info("Final metrics:")
	log.Info("  Events received: %d", snapshot.EventsReceived)
	log.Info("  Events processed: %d", snapshot.EventsProcessed)
	log.Info("  Events dropped: %d", snapshot.EventsDropped)
	log.Info("  Uptime: %d seconds", snapshot.UptimeSeconds)

	log.Info("Server stopped successfully")
}
