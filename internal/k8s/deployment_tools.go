package k8s

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// === Input/Output Types ===

// ListDeploymentsInput contains optional namespace
type ListDeploymentsInput struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"네임스페이스 (비어있으면 모든 네임스페이스)"`
}

// ListDeploymentsOutput contains list of deployments
type ListDeploymentsOutput struct {
	Deployments []DeploymentInfo `json:"deployments"`
}

// DeploymentStatusInput contains deployment identifiers
type DeploymentStatusInput struct {
	Namespace string `json:"namespace" jsonschema:"네임스페이스"`
	Name      string `json:"name" jsonschema:"Deployment 이름"`
}

// DeploymentStatusOutput contains detailed deployment status
type DeploymentStatusOutput struct {
	DeploymentInfo
}

// === Tool Handlers ===

func createListDeploymentsHandler(client *Client) mcp.ToolHandlerFor[ListDeploymentsInput, ListDeploymentsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListDeploymentsInput]) (*mcp.CallToolResultFor[ListDeploymentsOutput], error) {
		input := params.Arguments
		namespace := input.Namespace
		if namespace == "" {
			namespace = metav1.NamespaceAll
		}

		deployments, err := client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ListDeploymentsOutput]("Deployment 목록 조회 실패: %s", err.Error()), nil
		}

		output := ListDeploymentsOutput{
			Deployments: make([]DeploymentInfo, 0, len(deployments.Items)),
		}

		for _, deploy := range deployments.Items {
			strategy := string(deploy.Spec.Strategy.Type)
			var replicas int32 = 0
			if deploy.Spec.Replicas != nil {
				replicas = *deploy.Spec.Replicas
			}
			output.Deployments = append(output.Deployments, DeploymentInfo{
				Name:              deploy.Name,
				Namespace:         deploy.Namespace,
				Replicas:          replicas,
				ReadyReplicas:     deploy.Status.ReadyReplicas,
				AvailableReplicas: deploy.Status.AvailableReplicas,
				UpdatedReplicas:   deploy.Status.UpdatedReplicas,
				Strategy:          strategy,
				Labels:            deploy.Labels,
				CreationTime:      deploy.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createDeploymentStatusHandler(client *Client) mcp.ToolHandlerFor[DeploymentStatusInput, DeploymentStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeploymentStatusInput]) (*mcp.CallToolResultFor[DeploymentStatusOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[DeploymentStatusOutput]("namespace 파라미터가 필요합니다"), nil
		}
		if input.Name == "" {
			return toolErrorFor[DeploymentStatusOutput]("name 파라미터가 필요합니다"), nil
		}

		deploy, err := client.Clientset.AppsV1().Deployments(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
		if err != nil {
			return toolErrorfFor[DeploymentStatusOutput]("Deployment 조회 실패: %s", err.Error()), nil
		}

		strategy := string(deploy.Spec.Strategy.Type)
		var replicas int32 = 0
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}

		return toolSuccessFor(DeploymentStatusOutput{
			DeploymentInfo: DeploymentInfo{
				Name:              deploy.Name,
				Namespace:         deploy.Namespace,
				Replicas:          replicas,
				ReadyReplicas:     deploy.Status.ReadyReplicas,
				AvailableReplicas: deploy.Status.AvailableReplicas,
				UpdatedReplicas:   deploy.Status.UpdatedReplicas,
				Strategy:          strategy,
				Labels:            deploy.Labels,
				CreationTime:      deploy.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
			},
		}), nil
	}
}

// === Registration ===

func registerDeploymentTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_list_deployments",
		Description: "Deployment 목록을 조회합니다.",
	}, createListDeploymentsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_deployment_status",
		Description: "특정 Deployment의 상세 상태를 조회합니다.",
	}, createDeploymentStatusHandler(client))

	fmt.Println("Kubernetes deployment tools registered: k8s_list_deployments, k8s_deployment_status")
}
