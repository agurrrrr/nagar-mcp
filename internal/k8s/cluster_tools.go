package k8s

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// === Input/Output Types ===

// ClusterInfoInput is empty
type ClusterInfoInput struct{}

// ClusterInfoOutput contains cluster information
type ClusterInfoOutput struct {
	ClusterInfo
}

// === Tool Handlers ===

func createClusterInfoHandler(client *Client) mcp.ToolHandlerFor[ClusterInfoInput, ClusterInfoOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ClusterInfoInput]) (*mcp.CallToolResultFor[ClusterInfoOutput], error) {
		version, err := client.Clientset.Discovery().ServerVersion()
		if err != nil {
			return toolErrorfFor[ClusterInfoOutput]("failed to get cluster version: %s", err.Error()), nil
		}

		nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ClusterInfoOutput]("failed to list nodes: %s", err.Error()), nil
		}

		return toolSuccessFor(ClusterInfoOutput{
			ClusterInfo: ClusterInfo{
				Version:   version.GitVersion,
				Platform:  version.Platform,
				NodeCount: len(nodes.Items),
			},
		}), nil
	}
}

// === Registration ===

func registerClusterTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_cluster_info",
		Description: "Retrieves Kubernetes cluster information.",
	}, createClusterInfoHandler(client))

	slog.Info("Kubernetes cluster tools registered: k8s_cluster_info")
}
