package client

import "time"

type InstanceCreateParam struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	DeployProfile string                 `json:"deployProfile"`
	Version       string                 `json:"version"`
	Spec          SpecificationParam     `json:"spec"`
	Features      *InstanceFeatureParam  `json:"features,omitempty"`
	Tags          []TagParam             `json:"tags,omitempty"`
	ClusterId     string                 `json:"clusterId"`
	EndPoint      *InstanceEndpointParam `json:"endPoint,omitempty"`
	Accesses      []AccessCreateParam    `json:"accesses,omitempty"`
}

type SpecificationParam struct {
	ReservedAku          int32                      `json:"reservedAku"`
	NodeConfig           *NodeConfigParam           `json:"nodeConfig,omitempty"`
	Networks             []InstanceNetworkParam     `json:"networks,omitempty"`
	KubernetesNodeGroups []KubernetesNodeGroupParam `json:"kubernetesNodeGroups,omitempty"`
	SecurityGroup        *string                    `json:"securityGroup,omitempty"`
	BucketProfiles       []BucketProfileBindParam   `json:"bucketProfiles,omitempty"`
	Template             *string                    `json:"template,omitempty"`
}

type BucketProfileBindParam struct {
	Id *string `json:"id,omitempty"`
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
	WalMode         *string                `json:"walMode,omitempty"`
	Security        *InstanceSecurityParam `json:"security,omitempty"`
	Integrations    []IntegrationBindParam `json:"integrations,omitempty"`
	InstanceConfigs []ConfigItemParam      `json:"instanceConfigs,omitempty"`
}

type InstanceSecurityParam struct {
	AuthenticationMethods  []string `json:"authenticationMethods,omitempty"`
	TransitEncryptionModes []string `json:"transitEncryptionModes,omitempty"`
	CertificateAuthority   *string  `json:"certificateAuthority,omitempty"`
	CertificateChain       *string  `json:"certificateChain,omitempty"`
	PrivateKey             *string  `json:"privateKey,omitempty"`
	DataEncryptionMode     *string  `json:"dataEncryptionMode,omitempty"`
}

type IntegrationBindParam struct {
	Id *string `json:"id,omitempty"`
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
	NodeConfig           *NodeConfigVO            `json:"nodeConfig,omitempty"`
	ReservedAku          *int32                   `json:"reservedAku,omitempty"`
	CurrentAku           *int32                   `json:"currentAku,omitempty"`
	Networks             []InstanceZoneNetworkVO  `json:"networks,omitempty"`
	KubernetesNodeGroups []KubernetesNodeGroupVO  `json:"kubernetesNodeGroups,omitempty"`
	BucketProfiles       []BucketProfileSummaryVO `json:"bucketProfiles,omitempty"`
	SecurityGroupId      *string                  `json:"securityGroupId,omitempty"`
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

type InstanceFeatureVO struct {
	WalMode  *string                   `json:"walMode,omitempty"`
	Security *InstanceSecurityConfigVO `json:"security,omitempty"`
}

type InstanceSecurityConfigVO struct {
	AuthenticationMethods  []string `json:"authenticationMethods,omitempty"`
	TransitEncryptionModes []string `json:"transitEncryptionModes,omitempty"`
	DataEncryptionMode     *string  `json:"dataEncryptionMode,omitempty"`
}

type InstanceCertificateParam struct {
	CertificateAuthority string `json:"certificateAuthority"`
	CertificateChain     string `json:"certificateChain"`
	PrivateKey           string `json:"privateKey"`
}

type InstanceUpdateParam struct {
	Version *string                   `json:"version,omitempty"`
	Spec    *SpecificationUpdateParam `json:"spec,omitempty"`
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
	ReservedAku          int32                      `json:"reservedAku"`
	NodeConfig           *NodeConfigParam           `json:"nodeConfig,omitempty"`
	SecurityGroup        *string                    `json:"securityGroup,omitempty"`
	Template             *string                    `json:"template,omitempty"`
	Networks             []InstanceNetworkParam     `json:"networks,omitempty"`
	KubernetesNodeGroups []KubernetesNodeGroupParam `json:"kubernetesNodeGroups,omitempty"`
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
