package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Proxmox    ProxmoxConfig    `yaml:"proxmox"`
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
	OPNsense   OPNsenseConfig   `yaml:"opnsense"`
	Server     ServerConfig     `yaml:"server"`
}

// ProxmoxConfig holds Proxmox VE connection settings
type ProxmoxConfig struct {
	URL         string `yaml:"url"`
	TokenID     string `yaml:"token_id"`
	Secret      string `yaml:"secret"`
	InsecureTLS bool   `yaml:"insecure_tls"`
}

// KubernetesConfig holds Kubernetes connection settings
type KubernetesConfig struct {
	Kubeconfig string `yaml:"kubeconfig"`
	Context    string `yaml:"context"`
}

// OPNsenseConfig holds OPNsense connection settings
type OPNsenseConfig struct {
	URL         string `yaml:"url"`
	APIKey      string `yaml:"api_key"`
	APISecret   string `yaml:"api_secret"`
	InsecureTLS bool   `yaml:"insecure_tls"`
}

// ServerConfig holds MCP server settings
type ServerConfig struct {
	Transport string `yaml:"transport"`
	HTTPAddr  string `yaml:"http_addr"`
	Name      string `yaml:"name"`
	Version   string `yaml:"version"`
}

// Load reads and parses the configuration file, then applies environment variable overrides
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Expand home directory in paths
	cfg.expandPaths()

	// Set defaults
	cfg.setDefaults()

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to config
func (c *Config) applyEnvOverrides() {
	// Proxmox overrides
	if v := os.Getenv("PVE_URL"); v != "" {
		c.Proxmox.URL = v
	}
	if v := os.Getenv("PVE_TOKEN_ID"); v != "" {
		c.Proxmox.TokenID = v
	}
	if v := os.Getenv("PVE_SECRET"); v != "" {
		c.Proxmox.Secret = v
	}
	if v := os.Getenv("PVE_INSECURE_TLS"); v != "" {
		c.Proxmox.InsecureTLS = v == "true" || v == "1"
	}

	// OPNsense overrides
	if v := os.Getenv("OPS_URL"); v != "" {
		c.OPNsense.URL = v
	}
	if v := os.Getenv("OPS_API_KEY"); v != "" {
		c.OPNsense.APIKey = v
	}
	if v := os.Getenv("OPS_API_SECRET"); v != "" {
		c.OPNsense.APISecret = v
	}
	if v := os.Getenv("OPS_INSECURE_TLS"); v != "" {
		c.OPNsense.InsecureTLS = v == "true" || v == "1"
	}

	// Kubernetes overrides
	if v := os.Getenv("KUBECONFIG"); v != "" {
		c.Kubernetes.Kubeconfig = v
	}
	if v := os.Getenv("KUBE_CONTEXT"); v != "" {
		c.Kubernetes.Context = v
	}

	// Server overrides
	if v := os.Getenv("MCP_TRANSPORT"); v != "" {
		c.Server.Transport = v
	}
	if v := os.Getenv("MCP_HTTP_ADDR"); v != "" {
		c.Server.HTTPAddr = v
	}
}

// expandPaths expands ~ to home directory in path fields
func (c *Config) expandPaths() {
	c.Kubernetes.Kubeconfig = expandHomePath(c.Kubernetes.Kubeconfig)
}

// setDefaults sets default values for missing configuration
func (c *Config) setDefaults() {
	if c.Server.Transport == "" {
		c.Server.Transport = "stdio"
	}
	if c.Server.HTTPAddr == "" {
		c.Server.HTTPAddr = ":9090"
	}
	if c.Server.Name == "" {
		c.Server.Name = "nagar-mcp"
	}
	if c.Server.Version == "" {
		c.Server.Version = "0.1.0"
	}
}

// expandHomePath expands ~ to home directory
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Proxmox validation (optional - may not be configured)
	if c.Proxmox.URL != "" {
		if c.Proxmox.TokenID == "" {
			return fmt.Errorf("proxmox token_id is required when url is set")
		}
		if c.Proxmox.Secret == "" {
			return fmt.Errorf("proxmox secret is required when url is set")
		}
	}

	// OPNsense validation (optional)
	if c.OPNsense.URL != "" {
		if c.OPNsense.APIKey == "" {
			return fmt.Errorf("opnsense api_key is required when url is set")
		}
		if c.OPNsense.APISecret == "" {
			return fmt.Errorf("opnsense api_secret is required when url is set")
		}
	}

	// Server validation
	if c.Server.Transport != "stdio" && c.Server.Transport != "http" {
		return fmt.Errorf("invalid transport: %s (must be 'stdio' or 'http')", c.Server.Transport)
	}

	return nil
}

// HasProxmox returns true if Proxmox is configured
func (c *Config) HasProxmox() bool {
	return c.Proxmox.URL != "" && c.Proxmox.TokenID != "" && c.Proxmox.Secret != ""
}

// HasOPNsense returns true if OPNsense is configured
func (c *Config) HasOPNsense() bool {
	return c.OPNsense.URL != "" && c.OPNsense.APIKey != "" && c.OPNsense.APISecret != ""
}

// HasKubernetes returns true if Kubernetes is configured
func (c *Config) HasKubernetes() bool {
	// Kubernetes can work without explicit config (in-cluster mode)
	return true
}
