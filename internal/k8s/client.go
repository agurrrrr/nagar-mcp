package k8s

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"nagar-mcp/internal/config"
)

// Client wraps Kubernetes clients
type Client struct {
	Clientset       *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface
	config          *rest.Config
}

// NewClient creates a new Kubernetes client from configuration
func NewClient(cfg config.KubernetesConfig) (*Client, error) {
	var restConfig *rest.Config
	var err error

	if cfg.Kubeconfig != "" {
		// Build config from kubeconfig file
		loadingRules := &clientcmd.ClientConfigLoadingRules{
			ExplicitPath: cfg.Kubeconfig,
		}
		configOverrides := &clientcmd.ConfigOverrides{}
		if cfg.Context != "" {
			configOverrides.CurrentContext = cfg.Context
		}

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			configOverrides,
		)
		restConfig, err = kubeConfig.ClientConfig()
	} else {
		// Try in-cluster config
		restConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Create dynamic client
	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		Clientset:       clientset,
		DynamicClient:   dynClient,
		DiscoveryClient: clientset.Discovery(),
		config:          restConfig,
	}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *Client) IsConfigured() bool {
	return c != nil && c.Clientset != nil
}
