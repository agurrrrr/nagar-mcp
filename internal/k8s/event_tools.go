package k8s

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// === Input/Output Types ===

// ListEventsInput contains optional namespace and limit
type ListEventsInput struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"Namespace (all namespaces if empty)"`
	Limit     int64  `json:"limit,omitempty" jsonschema:"Maximum number of events (default 50)"`
}

// ListEventsOutput contains list of events
type ListEventsOutput struct {
	Events []EventInfo `json:"events"`
}

// === Tool Handlers ===

func createListEventsHandler(client *Client) mcp.ToolHandlerFor[ListEventsInput, ListEventsOutput] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListEventsInput]) (*mcp.CallToolResultFor[ListEventsOutput], error) {
		input := params.Arguments
		namespace := input.Namespace
		if namespace == "" {
			namespace = metav1.NamespaceAll
		}

		limit := input.Limit
		if limit == 0 {
			limit = 50
		}

		events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
			Limit: limit,
		})
		if err != nil {
			return toolErrorfFor[ListEventsOutput]("failed to list Events: %s", err.Error()), nil
		}

		output := ListEventsOutput{
			Events: make([]EventInfo, 0, len(events.Items)),
		}

		for _, event := range events.Items {
			involvedObject := fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name)

			output.Events = append(output.Events, EventInfo{
				Name:           event.Name,
				Namespace:      event.Namespace,
				Type:           event.Type,
				Reason:         event.Reason,
				Message:        event.Message,
				Count:          event.Count,
				FirstTimestamp: event.FirstTimestamp.Format("2006-01-02T15:04:05Z"),
				LastTimestamp:  event.LastTimestamp.Format("2006-01-02T15:04:05Z"),
				InvolvedObject: involvedObject,
			})
		}

		return toolSuccessFor(output), nil
	}
}

// === Registration ===

func registerEventTools(server *mcp.Server, client *Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "k8s_list_events",
		Description: "Lists Kubernetes events.",
	}, createListEventsHandler(client))

	slog.Info("Kubernetes event tools registered: k8s_list_events")
}
