package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"flight-event-throttler/internal/metrics"
	"flight-event-throttler/internal/model"
	"flight-event-throttler/pkg/logger"
)

// OpenSkyClient is a client for fetching data from OpenSky Network API
type OpenSkyClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	logger     *logger.Logger
	metrics    *metrics.Metrics
}

// NewOpenSkyClient creates a new OpenSky API client
func NewOpenSkyClient(baseURL string, timeout time.Duration, username, password string, log *logger.Logger, m *metrics.Metrics) *OpenSkyClient {
	return &OpenSkyClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		username: username,
		password: password,
		logger:   log,
		metrics:  m,
	}
}

// FetchAllStates fetches all current flight states from OpenSky API
func (c *OpenSkyClient) FetchAllStates(ctx context.Context) (*model.OpenSkyResponse, error) {
	url := fmt.Sprintf("%s/states/all", c.baseURL)
	return c.fetchStates(ctx, url)
}

// FetchStatesByBoundingBox fetches flight states within a geographic bounding box
func (c *OpenSkyClient) FetchStatesByBoundingBox(ctx context.Context, lamin, lomin, lamax, lomax float64) (*model.OpenSkyResponse, error) {
	url := fmt.Sprintf("%s/states/all?lamin=%.4f&lomin=%.4f&lamax=%.4f&lomax=%.4f",
		c.baseURL, lamin, lomin, lamax, lomax)
	return c.fetchStates(ctx, url)
}

// fetchStates is the internal method to fetch states from a given URL
func (c *OpenSkyClient) fetchStates(ctx context.Context, url string) (*model.OpenSkyResponse, error) {
	startTime := time.Now()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if credentials are provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "flight-event-throttler/1.0")

	// Increment API request metric
	if c.metrics != nil {
		c.metrics.IncrementAPIRequests()
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch data from OpenSky: %v", err)
		if c.metrics != nil {
			c.metrics.IncrementAPIErrors()
		}
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	// Record latency
	latency := time.Since(startTime).Milliseconds()
	if c.metrics != nil {
		c.metrics.RecordAPILatency(latency)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("OpenSky API returned status %d", resp.StatusCode)
		if c.metrics != nil {
			c.metrics.IncrementAPIErrors()
		}
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body: %v", err)
		if c.metrics != nil {
			c.metrics.IncrementAPIErrors()
		}
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var openSkyResp model.OpenSkyResponse
	if err := json.Unmarshal(body, &openSkyResp); err != nil {
		c.logger.Error("Failed to parse JSON response: %v", err)
		if c.metrics != nil {
			c.metrics.IncrementAPIErrors()
		}
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	c.logger.Debug("Fetched %d flight states from OpenSky API in %dms", len(openSkyResp.States), latency)

	return &openSkyResp, nil
}

// ConvertToFlightEvents converts OpenSky states to FlightEvent structs
func (c *OpenSkyClient) ConvertToFlightEvents(response *model.OpenSkyResponse) []*model.FlightEvent {
	if response == nil || len(response.States) == 0 {
		return nil
	}

	events := make([]*model.FlightEvent, 0, len(response.States))

	for _, state := range response.States {
		// OpenSky API returns state as array, need to map to struct
		// State format: [icao24, callsign, origin_country, time_position, last_contact,
		//                longitude, latitude, baro_altitude, on_ground, velocity,
		//                true_track, vertical_rate, sensors, geo_altitude, squawk, spi, position_source]

		if len(state) < 17 {
			c.logger.Debug("Skipping incomplete state data")
			continue
		}

		event := &model.FlightEvent{}

		// Extract ICAO24 (index 0)
		if icao24, ok := state[0].(string); ok {
			event.ICAO24 = icao24
		}

		// Extract Callsign (index 1)
		if callsign, ok := state[1].(string); ok {
			event.Callsign = callsign
		}

		// Extract Origin Country (index 2)
		if country, ok := state[2].(string); ok {
			event.OriginCountry = country
		}

		// Extract Time Position (index 3)
		if timePos, ok := state[3].(float64); ok {
			event.TimePosition = int64(timePos)
		}

		// Extract Last Contact (index 4)
		if lastContact, ok := state[4].(float64); ok {
			event.LastContact = int64(lastContact)
		}

		// Extract Longitude (index 5)
		if lon, ok := state[5].(float64); ok {
			event.Longitude = &lon
		}

		// Extract Latitude (index 6)
		if lat, ok := state[6].(float64); ok {
			event.Latitude = &lat
		}

		// Extract Baro Altitude (index 7)
		if alt, ok := state[7].(float64); ok {
			event.BaroAltitude = &alt
		}

		// Extract On Ground (index 8)
		if onGround, ok := state[8].(bool); ok {
			event.OnGround = onGround
		}

		// Extract Velocity (index 9)
		if vel, ok := state[9].(float64); ok {
			event.Velocity = &vel
		}

		// Extract True Track (index 10)
		if track, ok := state[10].(float64); ok {
			event.TrueTrack = &track
		}

		// Extract Vertical Rate (index 11)
		if vr, ok := state[11].(float64); ok {
			event.VerticalRate = &vr
		}

		// Extract Geo Altitude (index 13)
		if geoAlt, ok := state[13].(float64); ok {
			event.GeoAltitude = &geoAlt
		}

		// Extract Squawk (index 14)
		if squawk, ok := state[14].(string); ok {
			event.Squawk = &squawk
		}

		// Extract SPI (index 15)
		if spi, ok := state[15].(bool); ok {
			event.Spi = spi
		}

		// Extract Position Source (index 16)
		if posSource, ok := state[16].(float64); ok {
			event.PositionSource = int(posSource)
		}

		events = append(events, event)
	}

	c.logger.Debug("Converted %d OpenSky states to flight events", len(events))

	return events
}

// PollContinuously polls the OpenSky API at regular intervals
func (c *OpenSkyClient) PollContinuously(ctx context.Context, interval time.Duration, callback func([]*model.FlightEvent)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.logger.Info("Starting continuous polling of OpenSky API every %v", interval)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping OpenSky polling")
			return
		case <-ticker.C:
			response, err := c.FetchAllStates(ctx)
			if err != nil {
				c.logger.Error("Failed to fetch states during polling: %v", err)
				continue
			}

			events := c.ConvertToFlightEvents(response)
			if len(events) > 0 && callback != nil {
				callback(events)
			}
		}
	}
}
