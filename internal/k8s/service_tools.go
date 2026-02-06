package k8s

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// === Input/Output Types ===

// ListServicesInput contains optional namespace
type ListServicesInput struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"네임스페이스 (비어있으면 모든 네임스페이스)"`
}

// ListServicesOutput contains list of services
type ListServicesOutput struct {
	Services []ServiceInfo `json:"services"`
}

// ServiceStatusInput contains service identifiers
type ServiceStatusInput struct {
	Namespace string `json:"namespace" jsonschema:"네임스페이스"`
	Name      string `json:"name" jsonschema:"Service 이름"`
}

// ServiceStatusOutput contains detailed service status
type ServiceStatusOutput struct {
	ServiceInfo
}

// === Tool Handlers ===

func createListServicesHandler(client *Client) mcp.ToolHandlerFor[ListServicesInput, ListServicesOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListServicesInput]) (*mcp.CallToolResultFor[ListServicesOutput], error) {
		input := params.Arguments
		namespace := input.Namespace
		if namespace == "" {
			namespace = metav1.NamespaceAll
		}

		services, err := client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return toolErrorfFor[ListServicesOutput]("Service 목록 조회 실패: %s", err.Error()), nil
		}

		output := ListServicesOutput{
			Services: make([]ServiceInfo, 0, len(services.Items)),
		}

		for _, svc := range services.Items {
			ports := make([]ServicePort, 0, len(svc.Spec.Ports))
			for _, p := range svc.Spec.Ports {
				ports = append(ports, ServicePort{
					Name:       p.Name,
					Protocol:   string(p.Protocol),
					Port:       p.Port,
					TargetPort: p.TargetPort.String(),
					NodePort:   p.NodePort,
				})
			}

			output.Services = append(output.Services, ServiceInfo{
				Name:        svc.Name,
				Namespace:   svc.Namespace,
				Type:        string(svc.Spec.Type),
				ClusterIP:   svc.Spec.ClusterIP,
				ExternalIPs: svc.Spec.ExternalIPs,
				Ports:       ports,
				Selector:    svc.Spec.Selector,
			})
		}

		return toolSuccessFor(output), nil
	}
}

func createServiceStatusHandler(client *Client) mcp.ToolHandlerFor[ServiceStatusInput, ServiceStatusOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ServiceStatusInput]) (*mcp.CallToolResultFor[ServiceStatusOutput], error) {
		input := params.Arguments
		if input.Namespace == "" {
			return toolErrorFor[ServiceStatusOutput]("namespace 파라미터가 필요합니다"), nil
		}
		if input.Name == "" {
			return toolErrorFor[ServiceStatusOutput]("name 파라미터가 필요합니다"), nil
		}

		svc, err := client.Clientset.CoreV1().Services(input.Namespace).Get(ctx, input.Name, metav1.GetOptions{})
		if err != nil {
			return toolErrorfFor[ServiceStatusOutput]("Service 조회 실패: %s", err.Error()), nil
		}

		ports := make([]ServicePort, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			ports = append(ports, ServicePort{
				Name:       p.Name,
				Protocol:   string(p.Protocol),
				Port:       p.Port,
				TargetPort: p.TargetPort.String(),
				NodePort:   p.NodePort,
			})
		}

		return toolSuccessFor(ServiceStatusOutput{
			ServiceInfo: ServiceInfo{
				Name:        svc.Name,
				Namespace:   svc.Namespace,
				Type:        string(svc.Spec.Type),
				ClusterIP:   svc.Spec.ClusterIP,
				ExternalIPs: svc.Spec.ExternalIPs,
				Ports:       ports,
				Selector:    svc.Spec.Selector,
			},
		}), nil
	}
}

// === Registration ===

func registerServiceTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_list_services",
		Description: "Service 목록을 조회합니다.",
	}, createListServicesHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_service_status",
		Description: "특정 Service의 상세 정보를 조회합니다.",
	}, createServiceStatusHandler(client))

	fmt.Println("Kubernetes service tools registered: k8s_list_services, k8s_service_status")
}
