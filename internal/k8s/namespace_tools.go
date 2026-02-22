package k8s

import (
	"context"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// === Input/Output Types ===

// ListNamespacesInput is empty
type ListNamespacesInput struct{}

// ListNamespacesOutput contains list of namespaces
type ListNamespacesOutput struct {
	Namespaces []NamespaceInfo `json:"namespaces"`
}

// ListNodesInput is empty
type ListNodesInput struct{}

// ListNodesOutput contains list of nodes
type ListNodesOutput struct {
	Nodes []NodeInfo `json:"nodes"`
}

// NodeStatusInput contains the node name
type NodeStatusInput struct {
	Node string `json:"node" mcp:"description:Kubernetes node name"`
}

// NodeStatusOutput contains detailed node status
type NodeStatusOutput struct {
	NodeInfo
}

// === Tool Handlers ===

func createListNamespacesHandler(client *Client) mcp.ToolHandlerFor[ListNamespacesInput, ListNamespacesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNamespacesInput]) (*mcp.CallToolResultFor[ListNamespacesOutput], error) {
		namespaces, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ListNamespacesOutput]("failed to list namespaces: %s", err.Error()), nil
		}

		output := ListNamespacesOutput{
			Namespaces: make([]NamespaceInfo, 0, len(namespaces.Items)),
		}

		for _, ns := range namespaces.Items {
			output.Namespaces = append(output.Namespaces, NamespaceInfo{
				Name:         ns.Name,
				Status:       string(ns.Status.Phase),
				Labels:       ns.Labels,
				CreationTime: ns.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createListNodesHandler(client *Client) mcp.ToolHandlerFor[ListNodesInput, ListNodesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListNodesInput]) (*mcp.CallToolResultFor[ListNodesOutput], error) {
		nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ListNodesOutput]("failed to list nodes: %s", err.Error()), nil
		}

		output := ListNodesOutput{
			Nodes: make([]NodeInfo, 0, len(nodes.Items)),
		}

		for _, node := range nodes.Items {
			info := convertNodeToInfo(&node)
			output.Nodes = append(output.Nodes, info)
		}

		return toolSuccessFor(output), nil
	}
}

func createNodeStatusHandler(client *Client) mcp.ToolHandlerFor[NodeStatusInput, NodeStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[NodeStatusInput]) (*mcp.CallToolResultFor[NodeStatusOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[NodeStatusOutput]("node parameter is required"), nil
		}

		node, err := client.Clientset.CoreV1().Nodes().Get(ctx, input.Node, metav1.GetOptions{})
		if err != nil {
			return toolErrorfFor[NodeStatusOutput]("failed to get node: %s", err.Error()), nil
		}

		info := convertNodeToInfo(node)
		return toolSuccessFor(NodeStatusOutput{NodeInfo: info}), nil
	}
}

// Helper function to convert K8s Node to NodeInfo
func convertNodeToInfo(node *corev1.Node) NodeInfo {
	// Get node status
	status := "Unknown"
	conditions := make([]NodeCondition, 0)
	for _, cond := range node.Status.Conditions {
		conditions = append(conditions, NodeCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
		}
	}

	// Get roles
	roles := make([]string, 0)
	for label := range node.Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			if role != "" {
				roles = append(roles, role)
			}
		}
	}

	// Get IPs
	var internalIP, externalIP string
	for _, addr := range node.Status.Addresses {
		switch addr.Type {
		case corev1.NodeInternalIP:
			internalIP = addr.Address
		case corev1.NodeExternalIP:
			externalIP = addr.Address
		}
	}

	return NodeInfo{
		Name:              node.Name,
		Status:            status,
		Roles:             roles,
		Version:           node.Status.NodeInfo.KubeletVersion,
		InternalIP:        internalIP,
		ExternalIP:        externalIP,
		OS:                node.Status.NodeInfo.OSImage,
		Architecture:      node.Status.NodeInfo.Architecture,
		ContainerRuntime:  node.Status.NodeInfo.ContainerRuntimeVersion,
		AllocatableCPU:    node.Status.Allocatable.Cpu().String(),
		AllocatableMemory: node.Status.Allocatable.Memory().String(),
		Conditions:        conditions,
	}
}

// === Registration ===

func registerNamespaceTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_list_namespaces",
		Description: "Lists all namespaces in the Kubernetes cluster.",
	}, createListNamespacesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_list_nodes",
		Description: "Lists all nodes in the Kubernetes cluster.",
	}, createListNodesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_node_status",
		Description: "Retrieves the detailed status of a specific Kubernetes node.",
	}, createNodeStatusHandler(client))

	slog.Info("Kubernetes namespace/node tools registered: k8s_list_namespaces, k8s_list_nodes, k8s_node_status")
}
