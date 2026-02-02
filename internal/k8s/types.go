package k8s

// NodeInfo represents Kubernetes node information
type NodeInfo struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	Roles             []string          `json:"roles"`
	Version           string            `json:"version"`
	InternalIP        string            `json:"internal_ip,omitempty"`
	ExternalIP        string            `json:"external_ip,omitempty"`
	OS                string            `json:"os"`
	Architecture      string            `json:"architecture"`
	ContainerRuntime  string            `json:"container_runtime"`
	AllocatableCPU    string            `json:"allocatable_cpu"`
	AllocatableMemory string            `json:"allocatable_memory"`
	Conditions        []NodeCondition   `json:"conditions,omitempty"`
}

// NodeCondition represents a node condition
type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// PodInfo represents Kubernetes pod information
type PodInfo struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Status     string         `json:"status"`
	Phase      string         `json:"phase"`
	Node       string         `json:"node"`
	IP         string         `json:"ip,omitempty"`
	StartTime  string         `json:"start_time,omitempty"`
	Containers []K8sContainerInfo `json:"containers"`
	Restarts   int32          `json:"restarts"`
}

// K8sContainerInfo represents container information within a pod
type K8sContainerInfo struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"`
}

// DeploymentInfo represents Kubernetes deployment information
type DeploymentInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	UpdatedReplicas   int32             `json:"updated_replicas"`
	Strategy          string            `json:"strategy"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTime      string            `json:"creation_time"`
}

// ServiceInfo represents Kubernetes service information
type ServiceInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"cluster_ip"`
	ExternalIPs []string          `json:"external_ips,omitempty"`
	Ports       []ServicePort     `json:"ports"`
	Selector    map[string]string `json:"selector,omitempty"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	NodePort   int32  `json:"node_port,omitempty"`
}

// NamespaceInfo represents Kubernetes namespace information
type NamespaceInfo struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels,omitempty"`
	CreationTime string            `json:"creation_time"`
}

// EventInfo represents Kubernetes event information
type EventInfo struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Type           string `json:"type"`
	Reason         string `json:"reason"`
	Message        string `json:"message"`
	Count          int32  `json:"count"`
	FirstTimestamp string `json:"first_timestamp"`
	LastTimestamp  string `json:"last_timestamp"`
	InvolvedObject string `json:"involved_object"`
}

// IngressInfo represents Kubernetes ingress information
type IngressInfo struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Class       string        `json:"class,omitempty"`
	Hosts       []string      `json:"hosts"`
	Rules       []IngressRule `json:"rules"`
	TLS         []IngressTLS  `json:"tls,omitempty"`
}

// IngressRule represents an ingress rule
type IngressRule struct {
	Host  string        `json:"host"`
	Paths []IngressPath `json:"paths"`
}

// IngressPath represents an ingress path
type IngressPath struct {
	Path        string `json:"path"`
	PathType    string `json:"path_type"`
	ServiceName string `json:"service_name"`
	ServicePort string `json:"service_port"`
}

// IngressTLS represents TLS configuration for ingress
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secret_name"`
}

// ClusterInfo represents cluster information
type ClusterInfo struct {
	Version    string `json:"version"`
	Platform   string `json:"platform"`
	NodeCount  int    `json:"node_count"`
	APIServer  string `json:"api_server,omitempty"`
}
