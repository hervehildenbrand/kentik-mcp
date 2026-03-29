package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hervehildenbrand/kentik-mcp/internal/config"
)

type Client struct {
	email    string
	apiToken string
	v5Base   string
	v6Base   string
	http     *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		email:    cfg.Email,
		apiToken: cfg.APIToken,
		v5Base:   cfg.V5Base,
		v6Base:   cfg.V6Base,
		http: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) V5(method, path string, body any) (json.RawMessage, error) {
	return c.doRequest(method, c.v5Base+path, body)
}

func (c *Client) V6(method, path string, body any) (json.RawMessage, error) {
	result, err := c.doRequest(method, c.v6Base+path, body)
	if err != nil && isUnsupportedMediaType(err) {
		// Some V6 gRPC-gateway endpoints require grpc-web content type
		return c.doRequestWithContentType(method, c.v6Base+path, body, "application/grpc-web+json")
	}
	return result, err
}

func isUnsupportedMediaType(err error) bool {
	return err != nil && len(err.Error()) >= 13 && err.Error()[:13] == "API error 415"
}

func (c *Client) doRequestWithContentType(method, url string, body any, contentType string) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-CH-Auth-Email", c.email)
	req.Header.Set("X-CH-Auth-API-Token", c.apiToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return json.RawMessage(respBody), nil
}

func (c *Client) doRequest(method, url string, body any) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-CH-Auth-Email", c.email)
	req.Header.Set("X-CH-Auth-API-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return json.RawMessage(respBody), nil
}

// ListDeviceNames fetches all device names from the V5 API.
func (c *Client) ListDeviceNames() ([]string, error) {
	data, err := c.V5("GET", "/devices", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Devices []struct {
			DeviceName string `json:"device_name"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}
	names := make([]string, len(resp.Devices))
	for i, d := range resp.Devices {
		names[i] = d.DeviceName
	}
	return names, nil
}
