package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"nagar-mcp/internal/config"
	"nagar-mcp/internal/k8s"
	"nagar-mcp/internal/proxmox"
)

// Server wraps the MCP server with providers
type Server struct {
	mcpServer *mcp.Server
	config    *config.Config
	pveClient *proxmox.Client
	k8sClient *k8s.Client
}

// New creates a new MCP server with all providers
func New(cfg *config.Config) (*Server, error) {
	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	server := &Server{
		mcpServer: mcpServer,
		config:    cfg,
	}

	// Initialize Proxmox client if configured
	if cfg.HasProxmox() {
		pveClient, err := proxmox.NewClient(cfg.Proxmox)
		if err != nil {
			slog.Warn("failed to create Proxmox client", "error", err)
		} else {
			server.pveClient = pveClient
			slog.Info("Proxmox client initialized", "url", cfg.Proxmox.URL)
		}
	}

	// Initialize Kubernetes client if configured
	if cfg.HasKubernetes() {
		k8sClient, err := k8s.NewClient(cfg.Kubernetes)
		if err != nil {
			slog.Warn("failed to create Kubernetes client", "error", err)
		} else {
			server.k8sClient = k8sClient
			slog.Info("Kubernetes client initialized")
		}
	}

	// Register tools
	server.registerTools()

	return server, nil
}

// registerTools registers all MCP tools
func (s *Server) registerTools() {
	// Register Proxmox tools if client is available
	if s.pveClient != nil {
		proxmox.RegisterTools(s.mcpServer, s.pveClient)
		slog.Info("Proxmox tools registered")
	}

	// Register Kubernetes tools if client is available
	if s.k8sClient != nil {
		k8s.RegisterTools(s.mcpServer, s.k8sClient)
		slog.Info("Kubernetes tools registered")
	}
}

// Run starts the MCP server with the configured transport
func (s *Server) Run(ctx context.Context) error {
	switch s.config.Server.Transport {
	case "stdio":
		return s.runStdio(ctx)
	case "http":
		return s.runHTTP(ctx)
	default:
		return fmt.Errorf("unknown transport: %s", s.config.Server.Transport)
	}
}

// runStdio runs the server in stdio mode
func (s *Server) runStdio(ctx context.Context) error {
	slog.Info("starting MCP server in stdio mode")
	return s.mcpServer.Run(ctx, mcp.NewStdioTransport())
}

// runHTTP runs the server in HTTP mode
func (s *Server) runHTTP(ctx context.Context) error {
	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return s.mcpServer
		},
		nil,
	)

	http.Handle("/mcp", handler)

	slog.Info("starting MCP server in HTTP mode", "addr", s.config.Server.HTTPAddr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- http.ListenAndServe(s.config.Server.HTTPAddr, nil)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// MCPServer returns the underlying MCP server
func (s *Server) MCPServer() *mcp.Server {
	return s.mcpServer
}
