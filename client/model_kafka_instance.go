package client

import "time"

type InstanceCreateParam struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Spec        SpecificationParam     `json:"spec"`
	Features    *InstanceFeatureParam  `json:"features,omitempty"`
	Tags        []TagParam             `json:"tags,omitempty"`
	ClusterId   string                 `json:"clusterId,omitempty"`
	EndPoint    *InstanceEndpointParam `json:"endPoint,omitempty"`
	Accesses    []AccessCreateParam    `json:"accesses,omitempty"`
}

type SpecificationParam struct {
	ReservedAku              int32                      `json:"reservedAku"`
	NodeConfig               *NodeConfigParam           `json:"nodeConfig,omitempty"`
	Networks                 []InstanceNetworkParam     `json:"networks,omitempty"`
	KubernetesNodeGroups     []KubernetesNodeGroupParam `json:"kubernetesNodeGroups,omitempty"`
	SecurityGroup            *string                    `json:"securityGroup,omitempty"`
	Template                 *string                    `json:"template,omitempty"`
	FileSystem               *FileSystemParam           `json:"fileSystemForFsWal,omitempty"`
	DeployType               *string                    `json:"deployType,omitempty"`
	Provider                 *string                    `json:"provider,omitempty"`
	Region                   *string                    `json:"region,omitempty"`
	Scope                    *string                    `json:"scope,omitempty"`
	Vpc                      *string                    `json:"vpc,omitempty"`
	DnsZone                  *string                    `json:"dnsZone,omitempty"`
	KubernetesClusterId      *string                    `json:"kubernetesClusterId,omitempty"`
	KubernetesNamespace      *string                    `json:"kubernetesNamespace,omitempty"`
	KubernetesServiceAccount *string                    `json:"kubernetesServiceAccount,omitempty"`
	InstanceRole             *string                    `json:"instanceRole,omitempty"`
	DataBuckets              []BucketProfileParam       `json:"dataBuckets,omitempty"`
}

type BucketProfileParam struct {
	BucketName string  `json:"bucketName"`
	Provider   *string `json:"provider,omitempty"`
	Region     *string `json:"region,omitempty"`
	Scope      *string `json:"scope,omitempty"`
	Credential *string `json:"credential,omitempty"`
	Endpoint   *string `json:"endpoint,omitempty"`
}

type BucketProfileVO struct {
	Id          *string    `json:"id,omitempty"`
	BucketName  *string    `json:"bucketName,omitempty"`
	GmtCreate   *time.Time `json:"gmtCreate,omitempty"`
	GmtModified *time.Time `json:"gmtModified,omitempty"`
	Provider    *string    `json:"provider,omitempty"`
	Region      *string    `json:"region,omitempty"`
	Scope       *string    `json:"scope,omitempty"`
	Credential  *string    `json:"credential,omitempty"`
	Endpoint    *string    `json:"endpoint,omitempty"`
}

type FileSystemParam struct {
	ThroughputMiBpsPerFileSystem int32    `json:"throughputMiBpsPerFileSystem"`
	FileSystemCount              int32    `json:"fileSystemCount"`
	SecurityGroups               []string `json:"securityGroups,omitempty"`
}

type NodeConfigParam struct {
	PaymentType   *string  `json:"paymentType,omitempty"`
	PaymentPeriod *int32   `json:"paymentPeriod,omitempty"`
	InstanceTypes []string `json:"instanceTypes,omitempty"`
}

type InstanceNetworkParam struct {
	Zone   string  `json:"zone"`
	Subnet *string `json:"subnet,omitempty"`
}

type KubernetesNodeGroupParam struct {
	Id *string `json:"id,omitempty"`
}

type InstanceFeatureParam struct {
	WalMode         *string                       `json:"walMode,omitempty"`
	Security        *InstanceSecurityParam        `json:"security,omitempty"`
	InstanceConfigs []ConfigItemParam             `json:"instanceConfigs,omitempty"`
	S3Failover      *InstanceFailoverParam        `json:"s3Failover,omitempty"`
	MetricsExporter *InstanceMetricsExporterParam `json:"metricsExporter,omitempty"`
	TableTopic      *TableTopicParam              `json:"tableTopic,omitempty"`
	InboundRules    []InboundRuleParam            `json:"inboundRules,omitempty"`
	ExtendListeners []InstanceListenerParam       `json:"extendListeners,omitempty"`
}

type InstanceSecurityParam struct {
	AuthenticationMethods        []string `json:"authenticationMethods,omitempty"`
	TransitEncryptionModes       []string `json:"transitEncryptionModes,omitempty"`
	CertificateAuthority         *string  `json:"certificateAuthority,omitempty"`
	CertificateChain             *string  `json:"certificateChain,omitempty"`
	PrivateKey                   *string  `json:"privateKey,omitempty"`
	DataEncryptionMode           *string  `json:"dataEncryptionMode,omitempty"`
	TlsHostnameValidationEnabled *bool    `json:"tlsHostnameValidationEnabled,omitempty"`
}

type InstanceFailoverParam struct {
	Enabled           *bool   `json:"enabled,omitempty"`
	StorageType       *string `json:"storageType,omitempty"`
	EbsVolumeSizeInGB *int32  `json:"ebsVolumeSizeInGB,omitempty"`
}

type InstanceMetricsExporterParam struct {
	Prometheus *InstancePrometheusExporterParam `json:"prometheus,omitempty"`
	CloudWatch *InstanceCloudWatchExporterParam `json:"cloudWatch,omitempty"`
}

type InstancePrometheusExporterParam struct {
	Enabled       *bool               `json:"enabled,omitempty"`
	AuthType      *string             `json:"authType,omitempty"`
	EndPoint      *string             `json:"endPoint,omitempty"`
	PrometheusArn *string             `json:"prometheusArn,omitempty"`
	Username      *string             `json:"username,omitempty"`
	Password      *string             `json:"password,omitempty"`
	Token         *string             `json:"token,omitempty"`
	Labels        []MetricsLabelParam `json:"labels,omitempty"`
}

type InstanceCloudWatchExporterParam struct {
	Enabled   *bool   `json:"enabled,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
}

type MetricsLabelParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TableTopicParam struct {
	IntegrationId     *string `json:"integrationId,omitempty"`
	Warehouse         string  `json:"warehouse"`
	CatalogType       string  `json:"catalogType"`
	MetastoreUri      *string `json:"metastoreUri,omitempty"`
	HiveAuthMode      *string `json:"hiveAuthMode,omitempty"`
	KerberosPrincipal *string `json:"kerberosPrincipal,omitempty"`
	UserPrincipal     *string `json:"userPrincipal,omitempty"`
	KeytabFile        *string `json:"keytabFile,omitempty"`
	Krb5ConfFile      *string `json:"krb5confFile,omitempty"`
}

type InboundRuleParam struct {
	ListenerName string   `json:"listenerName"`
	Cidrs        []string `json:"cidrs"`
}

type InstanceListenerParam struct {
	ListenerName     string  `json:"listenerName"`
	SecurityProtocol *string `json:"securityProtocol,omitempty"`
	Port             *int32  `json:"port,omitempty"`
}

type TagParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type InstanceEndpointParam struct {
	Endpoint         string  `json:"endpoint"`
	SecurityProtocol *string `json:"securityProtocol,omitempty"`
	SaslMechanism    *string `json:"saslMechanism,omitempty"`
	User             *string `json:"user,omitempty"`
	Password         *string `json:"password,omitempty"`
}

type AccessCreateParam struct {
	Port *int32 `json:"port,omitempty"`
}

type InstanceSummaryVO struct {
	InstanceId     *string    `json:"instanceId,omitempty"`
	GmtCreate      *time.Time `json:"gmtCreate,omitempty"`
	GmtModified    *time.Time `json:"gmtModified,omitempty"`
	Name           *string    `json:"name,omitempty"`
	Description    *string    `json:"description,omitempty"`
	KafkaClusterId *string    `json:"kafkaClusterId,omitempty"`
	DeployProfile  *string    `json:"deployProfile,omitempty"`
	Version        *string    `json:"version,omitempty"`
	State          *string    `json:"state,omitempty"`
	Tags           []TagVO    `json:"tags,omitempty"`
}

type TagVO struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type InstanceVO struct {
	InstanceId     *string              `json:"instanceId,omitempty"`
	GmtCreate      *time.Time           `json:"gmtCreate,omitempty"`
	GmtModified    *time.Time           `json:"gmtModified,omitempty"`
	Name           *string              `json:"name,omitempty"`
	Description    *string              `json:"description,omitempty"`
	KafkaClusterId *string              `json:"kafkaClusterId,omitempty"`
	DeployProfile  *string              `json:"deployProfile,omitempty"`
	Version        *string              `json:"version,omitempty"`
	State          *string              `json:"state,omitempty"`
	Tags           []TagVO              `json:"tags,omitempty"`
	Statistics     *InstanceStatisticVO `json:"statistics,omitempty"`
	Spec           *SpecificationVO     `json:"spec,omitempty"`
	Features       *InstanceFeatureVO   `json:"features,omitempty"`
}

type InstanceStatisticVO struct {
	TopicCount          *int64   `json:"topicCount,omitempty"`
	PartitionCount      *int64   `json:"partitionCount,omitempty"`
	ConsumerGroupCount  *int64   `json:"consumerGroupCount,omitempty"`
	BytesInPerSecond    *float64 `json:"bytesInPerSecond,omitempty"`
	BytesOutPerSecond   *float64 `json:"bytesOutPerSecond,omitempty"`
	RequestsInPerSecond *float64 `json:"requestsInPerSecond,omitempty"`
	StorageSize         *int64   `json:"storageSize,omitempty"`
}

type SpecificationVO struct {
	NodeConfig               *NodeConfigVO            `json:"nodeConfig,omitempty"`
	ReservedAku              *int32                   `json:"reservedAku,omitempty"`
	CurrentAku               *int32                   `json:"currentAku,omitempty"`
	Networks                 []InstanceZoneNetworkVO  `json:"networks,omitempty"`
	KubernetesNodeGroups     []KubernetesNodeGroupVO  `json:"kubernetesNodeGroups,omitempty"`
	BucketProfiles           []BucketProfileSummaryVO `json:"bucketProfiles,omitempty"`
	DataBuckets              []BucketProfileVO        `json:"dataBuckets,omitempty"`
	SecurityGroupId          *string                  `json:"securityGroupId,omitempty"`
	FileSystem               *FileSystemVO            `json:"fileSystemForFsWal,omitempty"`
	Provider                 *string                  `json:"provider,omitempty"`
	Region                   *string                  `json:"region,omitempty"`
	Scope                    *string                  `json:"scope,omitempty"`
	Vpc                      *string                  `json:"vpc,omitempty"`
	DnsZone                  *string                  `json:"dnsZone,omitempty"`
	KubernetesClusterId      *string                  `json:"kubernetesClusterId,omitempty"`
	KubernetesNamespace      *string                  `json:"kubernetesNamespace,omitempty"`
	KubernetesServiceAccount *string                  `json:"kubernetesServiceAccount,omitempty"`
	InstanceRole             *string                  `json:"instanceRole,omitempty"`
	DeployType               *string                  `json:"deployType,omitempty"`
}

type NodeConfigVO struct {
	PaymentType   *string `json:"paymentType,omitempty"`
	PaymentPeriod *int32  `json:"paymentPeriod,omitempty"`
}

type InstanceZoneNetworkVO struct {
	Subnet     *string `json:"subnet,omitempty"`
	Zone       *string `json:"zone,omitempty"`
	SubnetName *string `json:"subnetName,omitempty"`
}

type KubernetesNodeGroupVO struct {
	Id *string `json:"id,omitempty"`
}

type BucketProfileSummaryVO struct {
	Id         *string `json:"id,omitempty"`
	BucketName *string `json:"bucketName,omitempty"`
}

type FileSystemVO struct {
	ThroughputMiBpsPerFileSystem *int32   `json:"throughputMiBpsPerFileSystem,omitempty"`
	FileSystemCount              *int32   `json:"fileSystemCount,omitempty"`
	SecurityGroups               []string `json:"securityGroups,omitempty"`
}

type InstanceFeatureVO struct {
	WalMode         *string                    `json:"walMode,omitempty"`
	Security        *InstanceSecurityConfigVO  `json:"security,omitempty"`
	S3Failover      *InstanceFailoverVO        `json:"s3Failover,omitempty"`
	MetricsExporter *InstanceMetricsExporterVO `json:"metricsExporter,omitempty"`
	TableTopic      *TableTopicVO              `json:"tableTopic,omitempty"`
	ExtendListeners []InstanceListenerVO       `json:"extendListeners,omitempty"`
	InboundRules    []InstanceInboundRuleVO    `json:"inboundRules,omitempty"`
}

type InstanceSecurityConfigVO struct {
	AuthenticationMethods        []string `json:"authenticationMethods,omitempty"`
	TransitEncryptionModes       []string `json:"transitEncryptionModes,omitempty"`
	DataEncryptionMode           *string  `json:"dataEncryptionMode,omitempty"`
	TlsHostnameValidationEnabled *bool    `json:"tlsHostnameValidationEnabled,omitempty"`
}

type InstanceFailoverVO struct {
	Enabled           bool    `json:"enabled"`
	StorageType       *string `json:"storageType,omitempty"`
	EbsVolumeSizeInGB *int32  `json:"ebsVolumeSizeInGB,omitempty"`
}

type InstanceMetricsExporterVO struct {
	Oltp       *InstanceOLTPExporterVO       `json:"oltp,omitempty"`
	Prometheus *InstancePrometheusExporterVO `json:"prometheus,omitempty"`
	CloudWatch *InstanceCloudWatchExporterVO `json:"cloudWatch,omitempty"`
}

type InstancePrometheusExporterVO struct {
	AuthType      *string          `json:"authType,omitempty"`
	EndPoint      *string          `json:"endPoint,omitempty"`
	Username      *string          `json:"username,omitempty"`
	PrometheusArn *string          `json:"prometheusArn,omitempty"`
	Labels        []MetricsLabelVO `json:"labels,omitempty"`
}

type MetricsLabelVO struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type InstanceCloudWatchExporterVO struct {
	Namespace *string `json:"namespace,omitempty"`
}

type TableTopicVO struct {
	Enabled           bool    `json:"enabled"`
	Warehouse         *string `json:"warehouse,omitempty"`
	CatalogType       *string `json:"catalogType,omitempty"`
	MetastoreUri      *string `json:"metastoreUri,omitempty"`
	HiveAuthMode      *string `json:"hiveAuthMode,omitempty"`
	KerberosPrincipal *string `json:"kerberosPrincipal,omitempty"`
	UserPrincipal     *string `json:"userPrincipal,omitempty"`
	KeytabFile        *string `json:"keytabFile,omitempty"`
	Krb5ConfFile      *string `json:"krb5confFile,omitempty"`
}

type InstanceListenerVO struct {
	ListenerName     *string `json:"listenerName,omitempty"`
	SecurityProtocol *string `json:"securityProtocol,omitempty"`
	Port             *int32  `json:"port,omitempty"`
}

type InstanceInboundRuleVO struct {
	ListenerName *string  `json:"listenerName,omitempty"`
	Cidrs        []string `json:"cidrs,omitempty"`
}

type InstanceOLTPExporterVO struct {
	EndPoint *string `json:"endPoint,omitempty"`
}

type InstanceCertificateParam struct {
	CertificateAuthority string `json:"certificateAuthority"`
	CertificateChain     string `json:"certificateChain"`
	PrivateKey           string `json:"privateKey"`
}

type InstanceUpdateParam struct {
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Version     *string                   `json:"version,omitempty"`
	Spec        *SpecificationUpdateParam `json:"spec,omitempty"`
	Features    *InstanceFeatureParam     `json:"features,omitempty"`
}

type InstanceBasicParam struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type InstanceVersionUpgradeParam struct {
	Version string `json:"version"`
}

type InstanceConfigParam struct {
	Configs []ConfigItemParam `json:"configs"`
}

type SpecificationUpdateParam struct {
	ReservedAku              *int32                     `json:"reservedAku,omitempty"`
	NodeConfig               *NodeConfigParam           `json:"nodeConfig,omitempty"`
	SecurityGroup            *string                    `json:"securityGroup,omitempty"`
	Template                 *string                    `json:"template,omitempty"`
	Networks                 []InstanceNetworkParam     `json:"networks,omitempty"`
	KubernetesNodeGroups     []KubernetesNodeGroupParam `json:"kubernetesNodeGroups,omitempty"`
	FileSystem               *FileSystemParam           `json:"fileSystemForFsWal,omitempty"`
	DeployType               *string                    `json:"deployType,omitempty"`
	Provider                 *string                    `json:"provider,omitempty"`
	Region                   *string                    `json:"region,omitempty"`
	Scope                    *string                    `json:"scope,omitempty"`
	Vpc                      *string                    `json:"vpc,omitempty"`
	DnsZone                  *string                    `json:"dnsZone,omitempty"`
	KubernetesClusterId      *string                    `json:"kubernetesClusterId,omitempty"`
	KubernetesNamespace      *string                    `json:"kubernetesNamespace,omitempty"`
	KubernetesServiceAccount *string                    `json:"kubernetesServiceAccount,omitempty"`
	InstanceRole             *string                    `json:"instanceRole,omitempty"`
	DataBuckets              []BucketProfileParam       `json:"dataBuckets,omitempty"`
}

type KafkaInstanceRequestPaymentPlan struct {
	PaymentType string `json:"paymentType"`
	Period      int    `json:"period"`
	Unit        string `json:"unit"`
}

type KafkaInstanceRequestNetwork struct {
	Zone   string `json:"zone"`
	Subnet string `json:"subnet"`
}

type KafkaInstanceResponse struct {
	InstanceID   string        `json:"instanceId"`
	GmtCreate    time.Time     `json:"gmtCreate"`
	GmtModified  time.Time     `json:"gmtModified"`
	DisplayName  string        `json:"displayName"`
	Description  string        `json:"description"`
	Status       string        `json:"status"`
	Provider     string        `json:"provider"`
	Region       string        `json:"region"`
	Spec         Spec          `json:"spec"`
	Networks     []Network     `json:"networks"`
	Metrics      []interface{} `json:"metrics"`
	AclSupported bool          `json:"aclSupported"`
	AclEnabled   bool          `json:"aclEnabled"`
}

type Spec struct {
	SpecID      string  `json:"specId"`
	DisplayName string  `json:"displayName"`
	Template    string  `json:"template"`
	Version     string  `json:"version"`
	Values      []Value `json:"currentValues"`
}

type Value struct {
	Key          string      `json:"key"`
	Name         string      `json:"name"`
	Value        interface{} `json:"value"`
	DisplayValue string      `json:"displayValue"`
}

type Network struct {
	Zone    string   `json:"zone"`
	Subnets []Subnet `json:"subnets"`
}

type Subnet struct {
	Subnet     string `json:"subnet"`
	SubnetName string `json:"subnetName"`
}

type PageNumResultInstanceVO struct {
	PageNum   *int32       `json:"pageNum,omitempty"`
	PageSize  *int32       `json:"pageSize,omitempty"`
	Total     *int64       `json:"total,omitempty"`
	List      []InstanceVO `json:"list,omitempty"`
	TotalPage *int64       `json:"totalPage,omitempty"`
}

type Metric struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Value       int    `json:"value"`
}

// PageNumResultInstanceAccessInfoVO struct for PageNumResultInstanceAccessInfoVO
type PageNumResultInstanceAccessInfoVO struct {
	List []InstanceAccessInfoVO `json:"list,omitempty"`
}

// InstanceAccessInfoVO struct for InstanceAccessInfoVO
type InstanceAccessInfoVO struct {
	DisplayName          *string         `json:"displayName,omitempty"`
	Name                 *string         `json:"name,omitempty"`
	NetworkType          *string         `json:"networkType,omitempty"`
	Protocol             *string         `json:"protocol,omitempty"`
	Mechanisms           *string         `json:"mechanisms,omitempty"`
	BootstrapServers     *string         `json:"bootstrapServers,omitempty"`
	LoadBalancerEndPoint *string         `json:"loadBalancerEndPoint,omitempty"`
	InboundRules         []InboundRuleVO `json:"inboundRules,omitempty"`
}

type InboundRuleVO struct {
	Cidr   *string `json:"cidr,omitempty"`
	System *bool   `json:"system,omitempty"`
}

// PageNumResultConfigItemVO struct for PageNumResultConfigItemVO
type PageNumResultConfigItemVO struct {
	PageNum   *int32            `json:"pageNum,omitempty"`
	PageSize  *int32            `json:"pageSize,omitempty"`
	Total     *int64            `json:"total,omitempty"`
	List      []ConfigItemParam `json:"list,omitempty"`
	TotalPage *int64            `json:"totalPage,omitempty"`
}
