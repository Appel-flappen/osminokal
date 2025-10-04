package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	defaultPollInterval string = "15m"
)

type Config struct {
	Source            SourceType
	CaCertPath        string
	ClientCertPath    string
	ClientCertKeyPath string
	Calendars         []CalendarConfig
	PollInterval      time.Duration
}

type CalendarConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func Load(logger *slog.Logger) (*Config, error) {
	cfg := &Config{}

	// 1. Load the simple, static variables
	sourceStr := os.Getenv("OSMINOKAL_SOURCE")
	cfg.Source = SourceType(strings.ToLower(sourceStr)) // Cast the loaded string to the custom type
	if !cfg.Source.IsValid() {
		return nil, fmt.Errorf("invalid source '%s'. Must be one of: %s",
			cfg.Source, AllowedSourcesString())
	}

	poll_str, ok := os.LookupEnv("OSMINOKAL_POLL_INTERVAL")
	if !ok {
		// use default interval of 15m
		poll_str = defaultPollInterval
	}
	dur, err := time.ParseDuration(poll_str)
	if err != nil {
		logger.Error("cannot parse poll interval duration, using default", "default", defaultPollInterval, "OSMINOKAL_POLL_INTERVAL", poll_str, "err", err)
		dur, _ = time.ParseDuration(defaultPollInterval)
	}

	cfg.PollInterval = dur

	cfg.CaCertPath = os.Getenv("OSMINOKAL_CA_CERT_PATH")
	cfg.ClientCertPath = os.Getenv("OSMINOKAL_CLIENT_CERT_PATH")
	cfg.ClientCertKeyPath = os.Getenv("OSMINOKAL_CLIENT_CERT_KEY_PATH")

	calendarJSON := os.Getenv("OSMINOKAL_CALENDARS")
	if calendarJSON == "" {
		return nil, fmt.Errorf("OSMINOKAL_CALENDARS environment variable is not set. Cannot proceed without any valid calendars")
	}
	var calendars []CalendarConfig
	if err := json.Unmarshal([]byte(calendarJSON), &calendars); err != nil {
		return nil, fmt.Errorf("failed to parse OSMINOKAL_CALENDARS json: %w", err)
	}
	for i := range calendars {
		calendars[i].Type = strings.ToLower(calendars[i].Type)
	}

	cfg.Calendars = calendars

	return cfg, nil
}
