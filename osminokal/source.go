package osminokal

import (
	"context"
	"fmt"
	"io"

	"net/http"

	"github.com/Appel-flappen/osminokal/config"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
)

type Source interface {
	GetSessions(ctx context.Context, deps Deps) ([]FreeEnergySession, error)
}

type SourceFactory func(c *http.Client, cfg *config.Config) (Source, error)

// New Source Location
var SourceRegistry = map[config.SourceType]SourceFactory{
	config.SourceWhizzy: NewWhizzyClient,
	config.SourceDavid:  NewDavidClient,
}

func NewSource(c *http.Client, cfg *config.Config) (Source, error) {
	factoryFunc, ok := SourceRegistry[cfg.Source]
	if !ok {
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Source)
	}

	return factoryFunc(c, cfg)
}

type WhizzyClient struct {
	HTTPClient *http.Client
	url        string
	name       string
}

func NewWhizzyClient(c *http.Client, cfg *config.Config) (Source, error) {
	return &WhizzyClient{
		HTTPClient: c,
		url:        "https://www.whizzy.org/octopus_powerups/free_electricity_session.json",
		name:       "Whizzy",
	}, nil
}

// GetSessions fetches free energy sessions from the whizzy api
func (c *WhizzyClient) GetSessions(ctx context.Context, deps Deps) ([]FreeEnergySession, error) {
	deps.Logger.Info("fetching sessions", "source", c.name)
	deps.Logger.Debug("fetching sessions", "url", c.url)
	resp, err := c.HTTPClient.Get(c.url)
	if err != nil {
		deps.Logger.Error("error making http request", "source", c.name, "err", err)
		return []FreeEnergySession{}, fmt.Errorf("couldn't get sessions from source")
	}

	var sessions []FreeEnergySession

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	err = nil
	if err != nil {
		deps.Logger.Error("error reading httpresponse", "source", c.name, "err", err)
		return sessions, fmt.Errorf("couldn't get sessions from source")
	}

	err = jsoniter.Unmarshal(body, &sessions)
	if err != nil {
		deps.Logger.Error("error unmarshaling json", "source", c.name, "err", err)
		return sessions, fmt.Errorf("couldn't get sessions from source")
	}

	// Whizzy api does not provide an ID, so they are made from start time in uuid v5 style.

	for i, _ := range sessions {
		name := []byte(sessions[i].Start.String())
		sessions[i].ID = uuid.NewSHA1(deps.NamespaceUUID, name)
	}

	return sessions, nil
}

type DavidClient struct {
	HTTPClient *http.Client
	url        string
	name       string
}

func NewDavidClient(c *http.Client, cfg *config.Config) (Source, error) {
	return &DavidClient{
		HTTPClient: c,
		url:        "https://oe-api.davidskendall.co.uk/free_electricity.json",
		name:       "David",
	}, nil
}

// GetSessions fetches free energy sessions from the david api
func (c *DavidClient) GetSessions(ctx context.Context, deps Deps) ([]FreeEnergySession, error) {
	resp, err := c.HTTPClient.Get(c.url)
	if err != nil {
		fmt.Println(err)
	}

	var sessions []FreeEnergySession
	fmt.Println(resp)

	// TODO implement logic

	return sessions, nil
}
