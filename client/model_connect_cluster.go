package client

import "time"

const (
	ConnectClusterStateCreating = "CREATING"
	ConnectClusterStateRunning  = "RUNNING"
	ConnectClusterStateFailed   = "FAILED"
	ConnectClusterStateChanging = "CHANGING"
	ConnectClusterStateDeleting = "DELETING"
	ConnectClusterStateDeleted  = "DELETED"
	ConnectClusterStateUnknown  = "UNKNOWN"
)

type ConnectClusterCreateParam struct {
	Name           string                      `json:"name"`
	Description    *string                     `json:"description,omitempty"`
	Plugins        []ClusterPluginParam        `json:"plugins"`
	KafkaCluster   ConnectClusterKafkaParam    `json:"kafkaCluster"`
	Capacity       ConnectClusterCapacityParam `json:"capacity"`
	Compute        ConnectClusterComputeParam  `json:"compute"`
	WorkerConfig   *ConnectorWorkerConfigParam `json:"workerConfig,omitempty"`
	MetricExporter *ConnectMetricsConfigParam  `json:"metricExporter,omitempty"`
	Tags           map[string]string           `json:"tags,omitempty"`
	Version        *string                     `json:"version,omitempty"`
}

type ConnectClusterUpdateParam struct {
	Name           *string                      `json:"name,omitempty"`
	Description    *string                      `json:"description,omitempty"`
	Plugins        []ClusterPluginParam         `json:"plugins,omitempty"`
	Capacity       *ConnectClusterCapacityParam `json:"capacity,omitempty"`
	WorkerConfig   *ConnectorWorkerConfigParam  `json:"workerConfig,omitempty"`
	MetricExporter *ConnectMetricsConfigParam   `json:"metricExporter,omitempty"`
	Tags           map[string]string            `json:"tags,omitempty"`
	Version        *string                      `json:"version,omitempty"`
}

type ClusterPluginParam struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ConnectClusterKafkaParam struct {
	KafkaInstanceId string `json:"kafkaInstanceId"`
}

type ConnectClusterCapacityParam struct {
	Type        string                          `json:"type"`
	Provisioned *ConnectClusterProvisionedParam `json:"provisioned,omitempty"`
	Autoscaling *ConnectClusterAutoscalingParam `json:"autoscaling,omitempty"`
}

type ConnectClusterProvisionedParam struct {
	WorkerResourceSpec string `json:"workerResourceSpec"`
	WorkerCount        int32  `json:"workerCount"`
}

type ConnectClusterAutoscalingParam struct {
	WorkerResourceSpec string            `json:"workerResourceSpec"`
	MinWorkerCount     int32             `json:"minWorkerCount"`
	MaxWorkerCount     int32             `json:"maxWorkerCount"`
	ScaleInPolicy      *ScalePolicyParam `json:"scaleInPolicy,omitempty"`
	ScaleOutPolicy     *ScalePolicyParam `json:"scaleOutPolicy,omitempty"`
}

type ScalePolicyParam struct {
	CpuUtilizationPercentage int32 `json:"cpuUtilizationPercentage"`
}

type ConnectClusterComputeParam struct {
	Type       string                  `json:"type"`
	Kubernetes *ConnectClusterK8sParam `json:"kubernetes,omitempty"`
	IamRole    *string                 `json:"iamRole,omitempty"`
}

type ConnectClusterK8sParam struct {
	ClusterId      string  `json:"clusterId"`
	Namespace      string  `json:"namespace"`
	ServiceAccount string  `json:"serviceAccount"`
	SchedulingSpec *string `json:"schedulingSpec,omitempty"`
}

type ConnectClusterVO struct {
	Id                       *string                 `json:"id,omitempty"`
	Name                     *string                 `json:"name,omitempty"`
	Description              *string                 `json:"description,omitempty"`
	State                    *string                 `json:"state,omitempty"`
	Plugins                  []ClusterPluginVO       `json:"plugins,omitempty"`
	KafkaInstanceId          *string                 `json:"kafkaInstanceId,omitempty"`
	KubernetesClusterId      *string                 `json:"kubernetesClusterId,omitempty"`
	KubernetesNamespace      *string                 `json:"kubernetesNamespace,omitempty"`
	KubernetesServiceAccount *string                 `json:"kubernetesServiceAccount,omitempty"`
	IamRole                  *string                 `json:"iamRole,omitempty"`
	WorkerCount              *int32                  `json:"workerCount,omitempty"`
	WorkerResourceSpec       *string                 `json:"workerResourceSpec,omitempty"`
	CapacityType             *string                 `json:"capacityType,omitempty"`
	MinWorkerCount           *int32                  `json:"minWorkerCount,omitempty"`
	MaxWorkerCount           *int32                  `json:"maxWorkerCount,omitempty"`
	ScaleInCpuPercent        *int32                  `json:"scaleInCpuPercent,omitempty"`
	ScaleOutCpuPercent       *int32                  `json:"scaleOutCpuPercent,omitempty"`
	WorkerConfig             map[string]interface{}  `json:"workerConfig,omitempty"`
	MetricExporter           *ConnectMetricsConfigVO `json:"metricExporter,omitempty"`
	Tags                     map[string]string       `json:"tags,omitempty"`
	Version                  *string                 `json:"version,omitempty"`
	KafkaConnectVersion      *string                 `json:"kafkaConnectVersion,omitempty"`
	SchedulingSpec           *string                 `json:"schedulingSpec,omitempty"`
	CreateTime               *time.Time              `json:"createTime,omitempty"`
	UpdateTime               *time.Time              `json:"updateTime,omitempty"`
}

type ClusterPluginVO struct {
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}
