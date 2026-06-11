// Package client provides an HTTP client that speaks to the LVA supervisor
// over its Unix domain socket at /run/lva/supervisor.sock.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	DefaultSocket  = "/run/lva/supervisor.sock"
	defaultTimeout = 30 * time.Second
	baseURL        = "http://lva"
)

// Client is a thin HTTP client wired to the supervisor socket.
type Client struct {
	http   *http.Client
	rawOut bool // when true, print raw JSON
}

// New returns a Client connected to the given socket path.
func New(socketPath string, rawOut bool) *Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: defaultTimeout}
			return d.DialContext(ctx, "unix", socketPath)
		},
	}
	return &Client{
		http: &http.Client{
			Transport: transport,
			Timeout:   defaultTimeout,
		},
		rawOut: rawOut,
	}
}

// APIResponse is the generic envelope the supervisor returns.
type APIResponse struct {
	// Raw body — for --raw-json passthrough or successful bodies.
	body []byte
}

// get sends a GET request and returns the body bytes or an error message.
func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.http.Get(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, apiError(body, resp.StatusCode)
	}
	return body, nil
}

// post sends a POST with an optional JSON body.
func (c *Client) post(path string, payload any) ([]byte, error) {
	var reqBody io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	var resp *http.Response
	var err error
	if reqBody != nil {
		resp, err = c.http.Post(baseURL+path, "application/json", reqBody)
	} else {
		resp, err = c.http.Post(baseURL+path, "application/json", http.NoBody)
	}
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, apiError(body, resp.StatusCode)
	}
	return body, nil
}

// apiError turns an error body into a readable error.
func apiError(body []byte, status int) error {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if msg, ok := obj["error"].(string); ok {
			return fmt.Errorf("supervisor error (%d): %s", status, msg)
		}
	}
	return fmt.Errorf("supervisor returned HTTP %d: %s", status, string(body))
}

// PrintJSON pretty-prints JSON, or raw if rawOut is set.
func (c *Client) PrintJSON(data []byte) error {
	if c.rawOut {
		fmt.Println(string(data))
		return nil
	}
	var pretty any
	if err := json.Unmarshal(data, &pretty); err != nil {
		// Not JSON — just print it.
		fmt.Println(string(data))
		return nil
	}
	out, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// =============================================================================
// API surface — one method per supervisor route
// =============================================================================

// --- containers ---

func (c *Client) ContainerList() ([]byte, error) {
	return c.get("/containers")
}

func (c *Client) ContainerStart(name string) ([]byte, error) {
	return c.post("/containers/"+name+"/start", nil)
}

func (c *Client) ContainerStop(name string) ([]byte, error) {
	return c.post("/containers/"+name+"/stop", nil)
}

func (c *Client) ContainerRestart(name string) ([]byte, error) {
	return c.post("/containers/"+name+"/restart", nil)
}

func (c *Client) ContainerUpdate(name string) ([]byte, error) {
	return c.post("/containers/"+name+"/update", nil)
}

func (c *Client) ContainerState(name string) ([]byte, error) {
	return c.get("/containers/" + name + "/state")
}

func (c *Client) ContainerStats(name string) ([]byte, error) {
	return c.get("/containers/" + name + "/stats")
}

func (c *Client) ContainerLogs(name string, tail int) ([]byte, error) {
	return c.get(fmt.Sprintf("/containers/%s/logs?tail=%d", name, tail))
}

// --- system ---

func (c *Client) SystemReboot() ([]byte, error) {
	return c.post("/system/reboot", nil)
}

func (c *Client) SystemPoweroff() ([]byte, error) {
	return c.post("/system/poweroff", nil)
}

func (c *Client) SystemOSUpdate(bundleURL string) ([]byte, error) {
	return c.post("/system/os-update", map[string]string{"bundle_url": bundleURL})
}

func (c *Client) SystemHealth() ([]byte, error) {
	return c.get("/system/health")
}

// --- network ---

func (c *Client) NetworkInfo() ([]byte, error) {
	return c.get("/network/info")
}

func (c *Client) NetworkInterfaces() ([]byte, error) {
	return c.get("/network/interfaces")
}

func (c *Client) NetworkSetHostname(hostname string) ([]byte, error) {
	return c.post("/network/hostname", map[string]string{"hostname": hostname})
}

func (c *Client) NetworkSetDHCP(iface string) ([]byte, error) {
	return c.post("/network/ip", map[string]string{
		"interface": iface,
		"method":    "dhcp",
	})
}

func (c *Client) NetworkSetStatic(iface, address string, prefix int, gateway string, dns []string) ([]byte, error) {
	return c.post("/network/ip", map[string]any{
		"interface": iface,
		"method":    "static",
		"address":   address,
		"prefix":    prefix,
		"gateway":   gateway,
		"dns":       dns,
	})
}

// --- wifi ---

func (c *Client) WifiScan(iface string) ([]byte, error) {
	return c.get("/network/wifi/scan?interface=" + iface)
}

func (c *Client) WifiConnect(iface, ssid string, password *string) ([]byte, error) {
	payload := map[string]any{
		"interface": iface,
		"ssid":      ssid,
	}
	if password != nil {
		payload["password"] = *password
	}
	return c.post("/network/wifi/connect", payload)
}

func (c *Client) WifiDisconnect(iface string) ([]byte, error) {
	return c.post("/network/wifi/disconnect", map[string]string{"interface": iface})
}

// --- updates ---

func (c *Client) UpdatesCheck() ([]byte, error) {
	return c.get("/updates")
}

func (c *Client) UpdatesVersions() ([]byte, error) {
	return c.get("/updates/versions")
}

func (c *Client) UpdatesOSInfo() ([]byte, error) {
	return c.get("/updates/os")
}

func (c *Client) UpdatesUpdateComponent(name string) ([]byte, error) {
	return c.post("/updates/update", map[string]string{"name": name})
}

// --- audio ---

func (c *Client) AudioDevices() ([]byte, error) {
	return c.get("/audio/devices")
}