package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Appel-flappen/osminokal/config"
	"github.com/Appel-flappen/osminokal/osminokal"
	"github.com/Appel-flappen/osminokal/version"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

type Deps struct {
	Logger *slog.Logger
}

// poll every x mins, default to 15 for the whizzy api

// if event exists:
// query calendar for event at same time with "made by osminokal"
// if event exists, do nothing
// else, create event and notify on ntfy

// SetupLogger reads the LOG_LEVEL environment variable and configures the default slog logger.
func SetupLogger() *slog.Logger {
	levelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	envLevel := os.Getenv("OSMINOKAL_LOG_LEVEL")
	envLevel = strings.ToLower(strings.TrimSpace(envLevel))

	level, ok := levelMap[envLevel]
	if !ok {
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(level)

	return logger
}

func SetupHttpClient(deps osminokal.Deps) *http.Client {
	RetryClient := retryablehttp.NewClient()
	RetryClient.RetryWaitMax = time.Second * 30
	RetryClient.RetryMax = 10
	RetryClient.Logger = deps.Logger

	return RetryClient.StandardClient()
}

func Process(ctx context.Context, deps osminokal.Deps, sc osminokal.Source, calendarList []osminokal.Calendar) {
	sessions, err := sc.GetSessions(ctx, deps)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			deps.Logger.Info("Received shutdown signal. Exiting", "signal", err)
			return
		}
		deps.Logger.Error("error getting sessions", "err", err)
		return
	}
	deps.Logger.Info(fmt.Sprintf("Successfully fetched %d sessions", len(sessions)))
	deps.Logger.Debug("session detail", "sessions", sessions)

	for _, cal := range calendarList {
		err = cal.PutSessions(deps, ctx, sessions)
		if err != nil {
			deps.Logger.Error("error submitting sessions", "err", err)
		} else {
			deps.Logger.Info("successfully put sessions to calendar")
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var deps osminokal.Deps

	logger := SetupLogger()
	logger.Info("Welcome to osminokal, friend ☭  🐙  🚬", "version", version.Version, "commit", version.Commit, "build_date", version.BuildDate)
	deps.Logger = logger
	deps.NamespaceUUID, _ = uuid.Parse("d0fb8c8a-8682-4c99-8365-333f3c727d63")

	// Get config
	cfg, err := config.Load(logger)
	if err != nil {
		logger.Error("critical: config cannot be loaded", "err", err)
		os.Exit(1)
	} else {
		logger.Info("config successfully loaded")
		logger.Debug("full config", "config", fmt.Sprintf("%#v", cfg))
	}

	// setup multi clients later, for now lets just use whizzy
	sourceHTTPClient := SetupHttpClient(deps)
	calendarHTTPClient := SetupHttpClient(deps)

	// create real source client
	sc, err := osminokal.NewSource(sourceHTTPClient, cfg)
	if err != nil {
		logger.Error("failed to create source client", "source", cfg.Source, "err", err)
	} else {
		logger.Info("created source client", "source", cfg.Source)
	}

	// create calendars based on config
	var calendarList []osminokal.Calendar

	for _, calCfg := range cfg.Calendars {
		newCal, err := osminokal.NewCalendar(deps, calendarHTTPClient, &calCfg)
		if err != nil {
			logger.Error("cannot create calendar", "name", calCfg.Name, "endpoint", calCfg.Endpoint, "err", err)
		}
		calendarList = append(calendarList, newCal)
	}

	// do properly
	// Setup certs if required
	if cfg.CaCertPath == "" || cfg.ClientCertKeyPath == "" || cfg.ClientCertPath == "" {
		logger.Info("one of CA cert path, Client cert path or Client cert key path not set, not setting up mTLS")
	} else if cfg.CaCertPath != "" && cfg.ClientCertKeyPath != "" && cfg.ClientCertPath != "" {
		caCert, err := os.ReadFile(cfg.CaCertPath)
		if err != nil {
			logger.Error("error reading ca certificate", "path", cfg.CaCertPath, "err", err)
		} else {
			logger.Info("loaded ca certificate", "path", cfg.CaCertPath)
		}
		caCertPool, _ := x509.SystemCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		cert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientCertKeyPath)
		if err != nil {
			logger.Error("error reading client certificate and key", "client_cert_path", cfg.ClientCertPath, "client_cert_key_path", cfg.ClientCertKeyPath, "err", err)
		} else {
			logger.Info("loaded client certificate and key")
			logger.Debug("loaded client certificate and key, debug info", "client_cert_path", cfg.ClientCertPath, "client_cert_key_path", cfg.ClientCertKeyPath)
		}

		calendarHTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		}
	}

	pollInterval := cfg.PollInterval

	logger.Info("started polling", "interval", pollInterval)

	// run initial oneshot
	Process(ctx, deps, sc, calendarList)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Received shutdown signal. Exiting")
			return
		case <-time.After(pollInterval):
			Process(ctx, deps, sc, calendarList)
		}
	}
}
