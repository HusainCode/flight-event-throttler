package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	OpenSky   OpenSkyConfig   `yaml:"opensky"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Buffer    BufferConfig    `yaml:"buffer"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type OpenSkyConfig struct {
	BaseURL       string        `yaml:"base_url"`
	PollInterval  time.Duration `yaml:"poll_interval"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	Username      string        `yaml:"username"`
	Password      string        `yaml:"password"`
}

type RateLimitConfig struct {
	EventsPerSecond int           `yaml:"events_per_second"`
	BurstSize       int           `yaml:"burst_size"`
	WindowDuration  time.Duration `yaml:"window_duration"`
}

type BufferConfig struct {
	Type       string `yaml:"type"` // "ring" or "sliding_window"
	Size       int    `yaml:"size"`
	BatchSize  int    `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}

type LoggingConfig struct {
	Level string `yaml:"level"` // "DEBUG", "INFO", "ERROR"
}

func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Set defaults
	config.setDefaults()

	// Load from file if provided
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	config.loadFromEnv()

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func (c *Config) setDefaults() {
	c.Server.Port = 8080
	c.Server.ReadTimeout = 15 * time.Second
	c.Server.WriteTimeout = 15 * time.Second
	c.Server.IdleTimeout = 60 * time.Second

	c.OpenSky.BaseURL = "https://opensky-network.org/api"
	c.OpenSky.PollInterval = 10 * time.Second
	c.OpenSky.RequestTimeout = 30 * time.Second

	c.RateLimit.EventsPerSecond = 100
	c.RateLimit.BurstSize = 200
	c.RateLimit.WindowDuration = 1 * time.Second

	c.Buffer.Type = "ring"
	c.Buffer.Size = 10000
	c.Buffer.BatchSize = 100
	c.Buffer.FlushInterval = 5 * time.Second

	c.Logging.Level = "INFO"
}

func (c *Config) loadFromEnv() {
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}

	if baseURL := os.Getenv("OPENSKY_BASE_URL"); baseURL != "" {
		c.OpenSky.BaseURL = baseURL
	}

	if username := os.Getenv("OPENSKY_USERNAME"); username != "" {
		c.OpenSky.Username = username
	}

	if password := os.Getenv("OPENSKY_PASSWORD"); password != "" {
		c.OpenSky.Password = password
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}

	if rps := os.Getenv("RATE_LIMIT_RPS"); rps != "" {
		if r, err := strconv.Atoi(rps); err == nil {
			c.RateLimit.EventsPerSecond = r
		}
	}

	if bufferType := os.Getenv("BUFFER_TYPE"); bufferType != "" {
		c.Buffer.Type = bufferType
	}

	if bufferSize := os.Getenv("BUFFER_SIZE"); bufferSize != "" {
		if s, err := strconv.Atoi(bufferSize); err == nil {
			c.Buffer.Size = s
		}
	}
}

func (c *Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	if c.OpenSky.BaseURL == "" {
		return fmt.Errorf("opensky base URL cannot be empty")
	}

	if c.RateLimit.EventsPerSecond < 1 {
		return fmt.Errorf("events per second must be at least 1")
	}

	if c.Buffer.Type != "ring" && c.Buffer.Type != "sliding_window" {
		return fmt.Errorf("buffer type must be 'ring' or 'sliding_window'")
	}

	if c.Buffer.Size < 1 {
		return fmt.Errorf("buffer size must be at least 1")
	}

	if c.Logging.Level != "DEBUG" && c.Logging.Level != "INFO" && c.Logging.Level != "ERROR" {
		return fmt.Errorf("log level must be 'DEBUG', 'INFO', or 'ERROR'")
	}

	return nil
}
