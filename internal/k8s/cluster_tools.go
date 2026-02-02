package k8s

import (
	"context"
	"fmt"

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
			return toolErrorfFor[ClusterInfoOutput]("클러스터 버전 조회 실패: %s", err.Error()), nil
		}

		nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ClusterInfoOutput]("노드 목록 조회 실패: %s", err.Error()), nil
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
		Description: "Kubernetes 클러스터 정보를 조회합니다.",
	}, createClusterInfoHandler(client))

	fmt.Println("Kubernetes cluster tools registered: k8s_cluster_info")
}
