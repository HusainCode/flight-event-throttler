# Flight Event Throttler

A high-performance Go application for fetching, throttling, and buffering real-time flight events from the OpenSky Network API. This service provides rate-limited event processing with configurable buffering strategies and comprehensive metrics collection.

## Features

- **Real-time Flight Data**: Continuously fetches flight data from OpenSky Network API
- **Rate Limiting**: Token bucket algorithm for precise event throttling
- **Flexible Buffering**: Choose between Ring Buffer or Sliding Window buffer implementations
- **Metrics Collection**: Comprehensive metrics tracking for monitoring system performance
- **RESTful API**: HTTP endpoints for retrieving events, metrics, and system stats
- **Graceful Shutdown**: Clean shutdown handling with proper resource cleanup
- **Configurable**: YAML configuration with environment variable overrides

## Architecture

```
┌─────────────────┐
│  OpenSky API    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  Fetcher Client │────→│ Rate Limiter │────→│   Buffer    │
└─────────────────┘     └──────────────┘     │ (Ring/SW)   │
                                              └──────┬──────┘
                                                     │
                                              ┌──────↓──────┐
                                              │  HTTP API   │
                                              └─────────────┘
```

## Project Structure

```
flight-event-throttler/
├── cmd/
│   └── server/
│       ├── main.go           # Application entry point
│       └── server.go         # Alternative server (unused)
├── internal/
│   ├── api/
│   │   └── http_server.go    # HTTP API handlers
│   ├── buffer/
│   │   ├── ring_buffer.go    # Circular buffer implementation
│   │   └── sliding_window.go # Sliding window buffer
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── fetcher/
│   │   └── opensky_client.go # OpenSky API client
│   ├── metrics/
│   │   └── metrics.go        # Metrics collection
│   ├── model/
│   │   └── event.go          # Data models
│   └── processor/
│       └── rate_limiter.go   # Rate limiting logic
├── pkg/
│   ├── logger/
│   │   └── logger.go         # Custom logger
│   └── utils/
│       └── time.go           # Time utilities
├── configs/
│   └── config.yaml           # Configuration file
├── scripts/
│   └── run_local.sh          # Local development script
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go 1.22.5 or higher
- Internet connection (for OpenSky API access)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd flight-event-throttler
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o bin/flight-event-throttler ./cmd/server
```

## Configuration

Configuration can be provided via YAML file or environment variables.

### Configuration Options

| Parameter | Environment Variable | Default | Description |
|-----------|---------------------|---------|-------------|
| `server.port` | `PORT` | `8080` | HTTP server port |
| `server.read_timeout` | - | `15s` | HTTP read timeout |
| `server.write_timeout` | - | `15s` | HTTP write timeout |
| `server.idle_timeout` | - | `60s` | HTTP idle timeout |
| `opensky.base_url` | `OPENSKY_BASE_URL` | `https://opensky-network.org/api` | OpenSky API base URL |
| `opensky.poll_interval` | - | `10s` | Polling interval |
| `opensky.username` | `OPENSKY_USERNAME` | - | OpenSky username (optional) |
| `opensky.password` | `OPENSKY_PASSWORD` | - | OpenSky password (optional) |
| `rate_limit.events_per_second` | `RATE_LIMIT_RPS` | `100` | Rate limit (events/sec) |
| `rate_limit.burst_size` | - | `200` | Burst size |
| `buffer.type` | `BUFFER_TYPE` | `ring` | Buffer type (`ring` or `sliding_window`) |
| `buffer.size` | `BUFFER_SIZE` | `10000` | Buffer capacity |
| `logging.level` | `LOG_LEVEL` | `INFO` | Log level (`DEBUG`, `INFO`, `ERROR`) |

### Example Configuration

```yaml
server:
  port: 8080
  read_timeout: 15s
  write_timeout: 15s

opensky:
  base_url: "https://opensky-network.org/api"
  poll_interval: 10s

rate_limit:
  events_per_second: 100
  burst_size: 200

buffer:
  type: "ring"
  size: 10000

logging:
  level: "INFO"
```

## Running the Application

### Using the Run Script

```bash
./scripts/run_local.sh
```

### Manual Execution

```bash
go run cmd/server/main.go
```

### With Environment Variables

```bash
PORT=9090 LOG_LEVEL=DEBUG BUFFER_TYPE=sliding_window go run cmd/server/main.go
```

## API Endpoints

### Health Check
```bash
GET /health
```

Returns the health status of the service.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1704067200,
  "uptime": "1h30m45s"
}
```

### Metrics
```bash
GET /metrics
```

Returns comprehensive system metrics.

**Response:**
```json
{
  "events_received": 15000,
  "events_processed": 14950,
  "events_dropped": 50,
  "events_failed": 0,
  "events_per_second": 98,
  "buffer_size": 9500,
  "buffer_capacity": 10000,
  "buffer_utilization_percent": 95.0,
  "api_requests": 150,
  "api_errors": 2,
  "api_avg_latency_ms": 245.5,
  "http_requests": 325,
  "http_errors": 0,
  "uptime_seconds": 5445,
  "timestamp": 1704067200
}
```

### Get All Events
```bash
GET /events
```

Returns all buffered flight events.

### Get Event Batch
```bash
GET /events/batch?size=100
```

Returns and removes a batch of events from the buffer.

**Query Parameters:**
- `size` (optional): Number of events to retrieve (default: 100)

### Buffer Statistics
```bash
GET /buffer/stats
```

Returns buffer statistics including count, capacity, and utilization.

## Buffer Types

### Ring Buffer
- Fixed-size circular buffer
- Overwrites oldest events when full
- Constant memory usage
- Best for: High-throughput scenarios where old data can be discarded

### Sliding Window Buffer
- Time-based event retention
- Automatically removes expired events
- Variable memory usage
- Best for: Time-sensitive applications requiring recent data

## Rate Limiting

The application uses a token bucket algorithm for rate limiting:

- **Events per Second**: Maximum sustained rate
- **Burst Size**: Maximum burst of events allowed
- **Window Duration**: Time window for rate calculations

Events exceeding the rate limit are dropped and counted in metrics.

## Metrics Tracking

The system tracks:
- **Event Metrics**: Received, processed, dropped, failed counts
- **Rate Metrics**: Events per second
- **Buffer Metrics**: Size, capacity, utilization percentage
- **API Metrics**: Request count, errors, average latency
- **HTTP Metrics**: Request count, errors
- **System Metrics**: Uptime

## Logging

Three log levels available:
- `DEBUG`: Detailed debugging information
- `INFO`: General information messages
- `ERROR`: Error messages only

Logs include timestamps and file locations for debugging.

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o flight-event-throttler ./cmd/server
```

## OpenSky Network API

This application uses the [OpenSky Network API](https://openskynetwork.github.io/opensky-api/) for real-time flight data.

- **Free Tier**: 400 requests/day, anonymous access
- **Authenticated**: Higher rate limits with account
- **Data**: Real-time aircraft positions, callsigns, velocities, and more

To use authenticated access, add credentials to config.yaml:
```yaml
opensky:
  username: "your_username"
  password: "your_password"
```

## Troubleshooting

### API Rate Limit Errors
- Reduce `poll_interval` in configuration
- Register for an OpenSky account for higher limits

### High Memory Usage
- Reduce `buffer.size` configuration
- Switch to `sliding_window` buffer type

### Events Being Dropped
- Increase `rate_limit.events_per_second`
- Increase `rate_limit.burst_size`
- Increase `buffer.size`

## Contributing

Contributions are welcome! Please submit pull requests or open issues for bugs and feature requests.

## Acknowledgments

- [OpenSky Network](https://opensky-network.org/) for providing free flight data API
- Go community for excellent libraries and tools
