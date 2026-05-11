package client

import "time"

// Connector lifecycle states returned by the API.
const (
	ConnectorStateCreating = "CREATING"
	ConnectorStateRunning  = "RUNNING"
	ConnectorStateFailed   = "FAILED"
	ConnectorStateChanging = "CHANGING"
	ConnectorStatePaused   = "PAUSED"
	ConnectorStateDeleting = "DELETING"
	ConnectorStateDeleted  = "DELETED"
	ConnectorStateUnknown  = "UNKNOWN"
)

// ---------------------------------------------------------------------------
// Request params
// ---------------------------------------------------------------------------

type ConnectorCreateParam struct {
	Name                     string                         `json:"name"`
	KubernetesClusterId      string                         `json:"kubernetesClusterId"`
	Description              *string                        `json:"description,omitempty"`
	PluginId                 string                         `json:"pluginId"`
	Capacity                 ConnectorCapacityParam         `json:"capacity"`
	TaskCount                int32                          `json:"taskCount"`
	Type                     *string                        `json:"type,omitempty"`
	ConnectorClass           *string                        `json:"connectorClass,omitempty"`
	IamRole                  *string                        `json:"iamRole,omitempty"`
	KubernetesServiceAccount string                         `json:"kubernetesServiceAccount"`
	KubernetesNamespace      string                         `json:"kubernetesNamespace"`
	KafkaCluster             ConnectorKafkaClusterParam     `json:"kafkaCluster"`
	WorkerConfig             *ConnectorWorkerConfigParam    `json:"workerConfig,omitempty"`
	ConnectorConfig          *ConnectorConnectorConfigParam `json:"connectorConfig,omitempty"`
	MetricExporter           *ConnectMetricsConfigParam     `json:"metricExporter,omitempty"`
	Version                  *string                        `json:"version,omitempty"`
	SchedulingSpec           *string                        `json:"schedulingSpec,omitempty"`
}

type ConnectorUpdateParam struct {
	Name                   *string                        `json:"name,omitempty"`
	Description            *string                        `json:"description,omitempty"`
	PluginId               *string                        `json:"pluginId,omitempty"`
	TaskCount              *int32                         `json:"taskCount,omitempty"`
	Capacity               *ConnectorCapacityParam        `json:"capacity,omitempty"`
	SecurityProtocolConfig *SecurityProtocolConfig        `json:"securityProtocolConfig,omitempty"`
	SchedulingSpec         *string                        `json:"schedulingSpec,omitempty"`
	WorkerConfig           *ConnectorWorkerConfigParam    `json:"workerConfig,omitempty"`
	ConnectorConfig        *ConnectorConnectorConfigParam `json:"connectorConfig,omitempty"`
	Labels                 map[string]string              `json:"labels,omitempty"`
	MetricExporter         *ConnectMetricsConfigParam     `json:"metricExporter,omitempty"`
	Version                *string                        `json:"version,omitempty"`
}

type ConnectorCapacityParam struct {
	WorkerCount        int32  `json:"workerCount"`
	WorkerResourceSpec string `json:"workerResourceSpec"`
}

type ConnectorKafkaClusterParam struct {
	KafkaInstanceId        string                 `json:"kafkaInstanceId"`
	SecurityProtocolConfig SecurityProtocolConfig `json:"securityProtocolConfig"`
}

type SecurityProtocolConfig struct {
	SecurityProtocol *string `json:"securityProtocol,omitempty"`
	Username         *string `json:"username,omitempty"`
	Password         *string `json:"password,omitempty"`
	SaslMechanism    *string `json:"saslMechanism,omitempty"`
	TruststorePath   *string `json:"truststorePath,omitempty"`
	TruststoreCerts  *string `json:"truststoreCerts,omitempty"`
	KeystorePath     *string `json:"keystorePath,omitempty"`
	KeyPassword      *string `json:"keyPassword,omitempty"`
	ClientCert       *string `json:"clientCert,omitempty"`
	PrivateKey       *string `json:"privateKey,omitempty"`
}

type ConnectorWorkerConfigParam struct {
	Properties map[string]string `json:"properties,omitempty"`
}

type ConnectorConnectorConfigParam struct {
	Properties map[string]string `json:"properties,omitempty"`
}

type ConnectMetricsConfigParam struct {
	RemoteWrite *ConnectRemoteWriteConfigParam `json:"remoteWrite,omitempty"`
}

type ConnectRemoteWriteConfigParam struct {
	Enabled            *bool             `json:"enabled,omitempty"`
	EndPoint           *string           `json:"endPoint,omitempty"`
	AuthType           *string           `json:"authType,omitempty"`
	Username           *string           `json:"username,omitempty"`
	Password           *string           `json:"password,omitempty"`
	Token              *string           `json:"token,omitempty"`
	Region             *string           `json:"region,omitempty"`
	PrometheusArn      *string           `json:"prometheusArn,omitempty"`
	InsecureSkipVerify *bool             `json:"insecureSkipVerify,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
}

// ---------------------------------------------------------------------------
// Response VOs
// ---------------------------------------------------------------------------

type ConnectorVO struct {
	Id                       *string                     `json:"id,omitempty"`
	Name                     *string                     `json:"name,omitempty"`
	Description              *string                     `json:"description,omitempty"`
	State                    *string                     `json:"state,omitempty"`
	WorkerCount              *int32                      `json:"workerCount,omitempty"`
	WorkerResourceSpec       *string                     `json:"workerResourceSpec,omitempty"`
	KafkaInstanceId          *string                     `json:"kafkaInstanceId,omitempty"`
	SecurityProtocolConfig   *SecurityProtocolConfig     `json:"securityProtocolConfig,omitempty"`
	KubernetesClusterId      *string                     `json:"kubernetesClusterId,omitempty"`
	KubernetesNamespace      *string                     `json:"kubernetesNamespace,omitempty"`
	KubernetesServiceAccount *string                     `json:"kubernetesServiceAccount,omitempty"`
	IamRole                  *string                     `json:"iamRole,omitempty"`
	Labels                   map[string]string           `json:"labels,omitempty"`
	ConnType                 *string                     `json:"connType,omitempty"`
	ConnClass                *string                     `json:"connClass,omitempty"`
	TaskCount                *int32                      `json:"taskCount,omitempty"`
	WorkerConfig             map[string]interface{}      `json:"workerConfig,omitempty"`
	ConnectorConfig          map[string]interface{}      `json:"connectorConfig,omitempty"`
	Plugin                   *ConnectPluginSummaryVO     `json:"plugin,omitempty"`
	MetricExporter           *ConnectMetricsConfigVO     `json:"metricExporter,omitempty"`
	Runtime                  *ConnectorRuntimeSnapshotVO `json:"runtime,omitempty"`
	CreateTime               *time.Time                  `json:"createTime,omitempty"`
	UpdateTime               *time.Time                  `json:"updateTime,omitempty"`
	Version                  *string                     `json:"version,omitempty"`
	KafkaConnectVersion      *string                     `json:"kafkaConnectVersion,omitempty"`
	SchedulingSpec           *string                     `json:"schedulingSpec,omitempty"`
	RenderedSchedulingSpec   *string                     `json:"renderedSchedulingSpec,omitempty"`
}

type ConnectPluginSummaryVO struct {
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type ConnectMetricsConfigVO struct {
	RemoteWrite *ConnectRemoteWriteConfigVO `json:"remoteWrite,omitempty"`
}

type ConnectRemoteWriteConfigVO struct {
	Enabled            bool              `json:"enabled"`
	EndPoint           *string           `json:"endPoint,omitempty"`
	AuthType           *string           `json:"authType,omitempty"`
	Username           *string           `json:"username,omitempty"`
	Region             *string           `json:"region,omitempty"`
	PrometheusArn      *string           `json:"prometheusArn,omitempty"`
	InsecureSkipVerify bool              `json:"insecureSkipVerify"`
	Headers            map[string]string `json:"headers,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
}

type ConnectorRuntimeSnapshotVO struct {
	State *string `json:"state,omitempty"`
}

type PageNumResultConnectorVO struct {
	PageNum   *int32        `json:"pageNum,omitempty"`
	PageSize  *int32        `json:"pageSize,omitempty"`
	Total     *int64        `json:"total,omitempty"`
	List      []ConnectorVO `json:"list,omitempty"`
	TotalPage *int64        `json:"totalPage,omitempty"`
}
