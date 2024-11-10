package truenas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	client *http.Client

	apiKey  string
	baseURL string
}

func NewClient(baseURL, apiKey string, client *http.Client) Client {
	if client == nil {
		client = &http.Client{}
	}
	return Client{
		client:  client,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// Apps gets running apps.
func (c Client) Apps() ([]App, error) {
	var apps []App
	err := c.do("/app", &apps)
	return apps, err
}

// Hostname gets the hostname of the TrueNAS system.
func (c Client) Hostname() (string, error) {
	var info systemInfo
	err := c.do("/system/info", &info)
	return info.Hostname, err
}

// Ping blocks until Pong is received or a timeout occurs.
func (c Client) Ping() error {
	var pong string
	err := c.do("/core/ping", &pong)
	if err != nil {
		return err
	}

	if pong != "pong" {
		return fmt.Errorf("unexpected response: %s", pong)
	}
	return nil
}

func (c Client) do(endpoint string, response any) error {
	if c.apiKey == "" {
		return errors.New("api key not set")
	}

	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New(resp.Status)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("unexpected content-type %s", resp.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	return nil
}

type App struct {
	Name    string  `json:"name"`
	State   State   `json:"state"`
	Portals Portals `json:"portals"`
}

type State string

const (
	StateCrashed   State = "CRASHED"
	StateDeploying       = "DEPLOYING"
	StateRunning         = "RUNNING"
	StateStopped         = "STOPPED"
	StateStopping        = "STOPPING"
)

type Portals []*url.URL

func (p *Portals) UnmarshalJSON(b []byte) error {
	untyped := make(map[string]string)
	err := json.Unmarshal(b, &untyped)
	if err != nil {
		return err
	}

	*p = make([]*url.URL, 0)
	for _, v := range untyped {
		u, err := url.Parse(v)
		if err != nil {
			return err
		}
		*p = append(*p, u)
	}

	return nil
}

type systemInfo struct {
	Hostname string
}
