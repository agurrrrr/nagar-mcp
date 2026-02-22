package k8s

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// === Input/Output Types ===

// ListPodsInput contains optional namespace
type ListPodsInput struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"Namespace (all namespaces if empty)"`
}

// ListPodsOutput contains list of pods
type ListPodsOutput struct {
	Pods []PodInfo `json:"pods"`
}

// PodStatusInput contains pod identifiers
type PodStatusInput struct {
	Namespace string `json:"namespace" jsonschema:"Namespace"`
	Name      string `json:"name" jsonschema:"Pod name"`
}

// PodStatusOutput contains detailed pod status
type PodStatusOutput struct {
	PodInfo
}

// PodLogsInput contains parameters for pod logs
type PodLogsInput struct {
	Namespace string `json:"namespace" jsonschema:"Namespace"`
	Name      string `json:"name" jsonschema:"Pod name"`
	Container string `json:"container,omitempty" jsonschema:"Container name (for multi-container Pods)"`
	TailLines int64  `json:"tail_lines,omitempty" jsonschema:"Number of lines to retrieve (default 100)"`
}

// PodLogsOutput contains pod logs
type PodLogsOutput struct {
	PodName   string `json:"pod_name"`
	Container string `json:"container,omitempty"`
	Logs      string `json:"logs"`
}

// === Tool Handlers ===

func createListPodsHandler(client *Client) mcp.ToolHandlerFor[ListPodsInput, ListPodsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListPodsInput]) (*mcp.CallToolResultFor[ListPodsOutput], error) {
		input := params.Arguments
		namespace := input.Namespace
		if namespace == "" {
			namespace = metav1.NamespaceAll
		}

		pods, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ListPodsOutput]("failed to list Pods: %s", err.Error()), nil
		}

		output := ListPodsOutput{
			Pods: make([]PodInfo, 0, len(pods.Items)),
		}

		for _, pod := range pods.Items {
			info := convertPodToInfo(&pod)
			output.Pods = append(output.Pods, info)
		}

		return toolSuccessFor(output), nil
	}
}

func createPodStatusHandler(client *Client) mcp.ToolHandlerFor[PodStatusInput, PodStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[PodStatusInput]) (*mcp.CallToolResultFor[PodStatusOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[PodStatusOutput]("namespace parameter is required"), nil
		}
		if input.Name == "" {
			return toolErrorFor[PodStatusOutput]("name parameter is required"), nil
		}

		pod, err := client.Clientset.CoreV1().Pods(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
		if err != nil {
			return toolErrorfFor[PodStatusOutput]("failed to get Pod: %s", err.Error()), nil
		}

		info := convertPodToInfo(pod)
		return toolSuccessFor(PodStatusOutput{PodInfo: info}), nil
	}
}

func createPodLogsHandler(client *Client) mcp.ToolHandlerFor[PodLogsInput, PodLogsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[PodLogsInput]) (*mcp.CallToolResultFor[PodLogsOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[PodLogsOutput]("namespace parameter is required"), nil
		}
		if input.Name == "" {
			return toolErrorFor[PodLogsOutput]("name parameter is required"), nil
		}

		tailLines := input.TailLines
		if tailLines == 0 {
			tailLines = 100
		}

		opts := &corev1.PodLogOptions{
			TailLines: &tailLines,
		}
		if input.Container != "" {
			opts.Container = input.Container
		}

		logs, err := client.Clientset.CoreV1().Pods(input.Namespace).GetLogs(input.Name, opts).Do(ctx).Raw()
		if err != nil {
			return toolErrorfFor[PodLogsOutput]("failed to get Pod logs: %s", err.Error()), nil
		}

		return toolSuccessFor(PodLogsOutput{
			PodName:   input.Name,
			Container: input.Container,
			Logs:      string(logs),
		}), nil
	}
}

// Helper function to convert K8s Pod to PodInfo
func convertPodToInfo(pod *corev1.Pod) PodInfo {
	containers := make([]K8sContainerInfo, 0, len(pod.Status.ContainerStatuses))
	var totalRestarts int32

	for _, cs := range pod.Status.ContainerStatuses {
		state := "Unknown"
		if cs.State.Running != nil {
			state = "Running"
		} else if cs.State.Waiting != nil {
			state = fmt.Sprintf("Waiting: %s", cs.State.Waiting.Reason)
		} else if cs.State.Terminated != nil {
			state = fmt.Sprintf("Terminated: %s", cs.State.Terminated.Reason)
		}

		containers = append(containers, K8sContainerInfo{
			Name:         cs.Name,
			Image:        cs.Image,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			State:        state,
		})
		totalRestarts += cs.RestartCount
	}

	var startTime string
	if pod.Status.StartTime != nil {
		startTime = pod.Status.StartTime.Format("2006-01-02T15:04:05Z")
	}

	return PodInfo{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     string(pod.Status.Phase),
		Phase:      string(pod.Status.Phase),
		Node:       pod.Spec.NodeName,
		IP:         pod.Status.PodIP,
		StartTime:  startTime,
		Containers: containers,
		Restarts:   totalRestarts,
	}
}

// === Registration ===

func registerPodTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_list_pods",
		Description: "Lists Pods in a specific namespace or across all namespaces.",
	}, createListPodsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_pod_status",
		Description: "Retrieves the detailed status of a specific Pod.",
	}, createPodStatusHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_pod_logs",
		Description: "Retrieves Pod logs. A specific container can be specified.",
	}, createPodLogsHandler(client))

	slog.Info("Kubernetes pod tools registered: k8s_list_pods, k8s_pod_status, k8s_pod_logs")
}
