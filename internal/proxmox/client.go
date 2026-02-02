package proxmox

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/luthermonson/go-proxmox"

	"nagar-mcp/internal/config"
)

// Client wraps the go-proxmox client with additional functionality
type Client struct {
	*proxmox.Client
	config *config.ProxmoxConfig
}

// NewClient creates a new Proxmox client from configuration
func NewClient(cfg config.ProxmoxConfig) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("proxmox URL is required")
	}
	if cfg.TokenID == "" {
		return nil, fmt.Errorf("proxmox token_id is required")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("proxmox secret is required")
	}

	opts := []proxmox.Option{
		proxmox.WithAPIToken(cfg.TokenID, cfg.Secret),
	}

	// Configure TLS if insecure mode is enabled
	if cfg.InsecureTLS {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		opts = append(opts, proxmox.WithHTTPClient(httpClient))
	}

	client := proxmox.NewClient(cfg.URL, opts...)

	return &Client{
		Client: client,
		config: &cfg,
	}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *Client) IsConfigured() bool {
	return c != nil && c.Client != nil
}
