package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// === Input/Output Types for Write Operations ===

// ScaleDeploymentInput contains parameters for scaling a deployment
type ScaleDeploymentInput struct {
	Namespace string `json:"namespace" jsonschema:"Namespace"`
	Name      string `json:"name" jsonschema:"Deployment name"`
	Replicas  int32  `json:"replicas" jsonschema:"Desired number of replicas"`
}

// ScaleDeploymentOutput contains the result of scaling
type ScaleDeploymentOutput struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	PreviousReplicas int32 `json:"previous_replicas"`
	DesiredReplicas int32  `json:"desired_replicas"`
	Message         string `json:"message"`
}

// RestartDeploymentInput contains parameters for restarting a deployment
type RestartDeploymentInput struct {
	Namespace string `json:"namespace" jsonschema:"Namespace"`
	Name      string `json:"name" jsonschema:"Deployment name"`
}

// RestartDeploymentOutput contains the result of restarting
type RestartDeploymentOutput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Message   string `json:"message"`
}

// DeletePodInput contains parameters for deleting a pod
type DeletePodInput struct {
	Namespace string `json:"namespace" jsonschema:"Namespace"`
	Name      string `json:"name" jsonschema:"Pod name"`
	Force     bool   `json:"force,omitempty" jsonschema:"Force delete (GracePeriod=0)"`
}

// DeletePodOutput contains the result of deleting a pod
type DeletePodOutput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Message   string `json:"message"`
}

// CreateNamespaceInput contains parameters for creating a namespace
type CreateNamespaceInput struct {
	Name   string            `json:"name" jsonschema:"Namespace name"`
	Labels map[string]string `json:"labels,omitempty" jsonschema:"Labels"`
}

// CreateNamespaceOutput contains the result of creating a namespace
type CreateNamespaceOutput struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// DeleteNamespaceInput contains parameters for deleting a namespace
type DeleteNamespaceInput struct {
	Name string `json:"name" jsonschema:"Namespace name"`
}

// DeleteNamespaceOutput contains the result of deleting a namespace
type DeleteNamespaceOutput struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// CordonNodeInput contains parameters for cordoning a node
type CordonNodeInput struct {
	Node string `json:"node" jsonschema:"Node name"`
}

// CordonNodeOutput contains the result of cordoning
type CordonNodeOutput struct {
	Node    string `json:"node"`
	Message string `json:"message"`
}

// UncordonNodeInput contains parameters for uncordoning a node
type UncordonNodeInput struct {
	Node string `json:"node" jsonschema:"Node name"`
}

// UncordonNodeOutput contains the result of uncordoning
type UncordonNodeOutput struct {
	Node    string `json:"node"`
	Message string `json:"message"`
}

// === Tool Handlers ===

func createScaleDeploymentHandler(client *Client) mcp.ToolHandlerFor[ScaleDeploymentInput, ScaleDeploymentOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ScaleDeploymentInput]) (*mcp.CallToolResultFor[ScaleDeploymentOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[ScaleDeploymentOutput]("namespace parameter is required"), nil
		}
		if input.Name == "" {
			return toolErrorFor[ScaleDeploymentOutput]("name parameter is required"), nil
		}
		if input.Replicas < 0 {
			return toolErrorFor[ScaleDeploymentOutput]("replicas must be 0 or greater"), nil
		}

		// Get current deployment
		deploy, err := client.Clientset.AppsV1().Deployments(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
		if err != nil {
			return toolErrorfFor[ScaleDeploymentOutput]("failed to query deployment: %s", err.Error()), nil
		}

		var previousReplicas int32 = 0
		if deploy.Spec.Replicas != nil {
			previousReplicas = *deploy.Spec.Replicas
		}

		// Scale the deployment
		scale, err := client.Clientset.AppsV1().Deployments(input.Namespace).GetScale(ctx, input.Name, metav1.GetOptions{})
		if err != nil {
			return toolErrorfFor[ScaleDeploymentOutput]("failed to get scale: %s", err.Error()), nil
		}

		scale.Spec.Replicas = input.Replicas
		_, err = client.Clientset.AppsV1().Deployments(input.Namespace).UpdateScale(ctx, input.Name, scale, metav1.UpdateOptions{})
		if err != nil {
			return toolErrorfFor[ScaleDeploymentOutput]("failed to update scale: %s", err.Error()), nil
		}

		return toolSuccessFor(ScaleDeploymentOutput{
			Name:             input.Name,
			Namespace:        input.Namespace,
			PreviousReplicas: previousReplicas,
			DesiredReplicas:  input.Replicas,
			Message:          fmt.Sprintf("Deployment %s/%s scaled from %d to %d replicas", input.Namespace, input.Name, previousReplicas, input.Replicas),
		}), nil
	}
}

func createRestartDeploymentHandler(client *Client) mcp.ToolHandlerFor[RestartDeploymentInput, RestartDeploymentOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[RestartDeploymentInput]) (*mcp.CallToolResultFor[RestartDeploymentOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[RestartDeploymentOutput]("namespace parameter is required"), nil
		}
		if input.Name == "" {
			return toolErrorFor[RestartDeploymentOutput]("name parameter is required"), nil
		}

		// Patch the deployment with a restart annotation (same as kubectl rollout restart)
		patch := map[string]interface{}{
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
						},
					},
				},
			},
		}

		patchBytes, err := json.Marshal(patch)
		if err != nil {
			return toolErrorfFor[RestartDeploymentOutput]("failed to create patch: %s", err.Error()), nil
		}

		_, err = client.Clientset.AppsV1().Deployments(input.Namespace).Patch(
			ctx,
			input.Name,
			types.StrategicMergePatchType,
			patchBytes,
			metav1.PatchOptions{},
		)
		if err != nil {
			return toolErrorfFor[RestartDeploymentOutput]("failed to restart deployment: %s", err.Error()), nil
		}

		return toolSuccessFor(RestartDeploymentOutput{
			Name:      input.Name,
			Namespace: input.Namespace,
			Message:   fmt.Sprintf("Deployment %s/%s restarted (rollout restart triggered)", input.Namespace, input.Name),
		}), nil
	}
}

func createDeletePodHandler(client *Client) mcp.ToolHandlerFor[DeletePodInput, DeletePodOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeletePodInput]) (*mcp.CallToolResultFor[DeletePodOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[DeletePodOutput]("namespace parameter is required"), nil
		}
		if input.Name == "" {
			return toolErrorFor[DeletePodOutput]("name parameter is required"), nil
		}

		deleteOptions := metav1.DeleteOptions{}
		if input.Force {
			gracePeriod := int64(0)
			deleteOptions.GracePeriodSeconds = &gracePeriod
		}

		err := client.Clientset.CoreV1().Pods(input.Namespace).Delete(ctx, input.Name, deleteOptions)
		if err != nil {
			return toolErrorfFor[DeletePodOutput]("failed to delete pod: %s", err.Error()), nil
		}

		forceMsg := ""
		if input.Force {
			forceMsg = " (force delete)"
		}

		return toolSuccessFor(DeletePodOutput{
			Name:      input.Name,
			Namespace: input.Namespace,
			Message:   fmt.Sprintf("Pod %s/%s deleted%s", input.Namespace, input.Name, forceMsg),
		}), nil
	}
}

func createCordonNodeHandler(client *Client) mcp.ToolHandlerFor[CordonNodeInput, CordonNodeOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CordonNodeInput]) (*mcp.CallToolResultFor[CordonNodeOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[CordonNodeOutput]("node parameter is required"), nil
		}

		// Patch the node to set unschedulable = true
		patch := []byte(`{"spec":{"unschedulable":true}}`)
		_, err := client.Clientset.CoreV1().Nodes().Patch(
			ctx,
			input.Node,
			types.StrategicMergePatchType,
			patch,
			metav1.PatchOptions{},
		)
		if err != nil {
			return toolErrorfFor[CordonNodeOutput]("failed to cordon node: %s", err.Error()), nil
		}

		return toolSuccessFor(CordonNodeOutput{
			Node:    input.Node,
			Message: fmt.Sprintf("Node %s cordoned (marked as unschedulable)", input.Node),
		}), nil
	}
}

func createUncordonNodeHandler(client *Client) mcp.ToolHandlerFor[UncordonNodeInput, UncordonNodeOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[UncordonNodeInput]) (*mcp.CallToolResultFor[UncordonNodeOutput], error) {
		input := params.Arguments
		if input.Node == "" {
			return toolErrorFor[UncordonNodeOutput]("node parameter is required"), nil
		}

		// Patch the node to set unschedulable = false
		patch := []byte(`{"spec":{"unschedulable":false}}`)
		_, err := client.Clientset.CoreV1().Nodes().Patch(
			ctx,
			input.Node,
			types.StrategicMergePatchType,
			patch,
			metav1.PatchOptions{},
		)
		if err != nil {
			return toolErrorfFor[UncordonNodeOutput]("failed to uncordon node: %s", err.Error()), nil
		}

		return toolSuccessFor(UncordonNodeOutput{
			Node:    input.Node,
			Message: fmt.Sprintf("Node %s uncordoned (marked as schedulable)", input.Node),
		}), nil
	}
}

// === Registration ===

func RegisterActionTools(server *mcp.Server, client *Client) {
	// k8s_scale_deployment
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_scale_deployment",
		Description: "Scales the number of replicas for a deployment.",
	}, createScaleDeploymentHandler(client))

	// k8s_restart_deployment
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_restart_deployment",
		Description: "Restarts a deployment (rollout restart).",
	}, createRestartDeploymentHandler(client))

	// k8s_delete_pod
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_delete_pod",
		Description: "Deletes a pod. Force delete available with force option.",
	}, createDeletePodHandler(client))

	// k8s_cordon_node
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_cordon_node",
		Description: "Cordons a node to prevent new pods from being scheduled.",
	}, createCordonNodeHandler(client))

	// k8s_uncordon_node
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_uncordon_node",
		Description: "Uncordons a node to allow pod scheduling again.",
	}, createUncordonNodeHandler(client))

	slog.Info("Kubernetes action tools registered: k8s_scale_deployment, k8s_restart_deployment, k8s_delete_pod, k8s_cordon_node, k8s_uncordon_node")
}
