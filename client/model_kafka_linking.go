package client

import "time"

// KafkaLinkSourceClusterParam represents inline source cluster configuration for creating links.
type KafkaLinkSourceClusterParam struct {
	Endpoint                      string  `json:"endpoint"`
	SecurityProtocol              *string `json:"securityProtocol,omitempty"`
	SaslMechanism                 *string `json:"saslMechanism,omitempty"`
	User                          *string `json:"user,omitempty"`
	Password                      *string `json:"password,omitempty"`
	TruststoreCertificates        *string `json:"truststoreCertificates,omitempty"`
	KeystoreCertificateChain      *string `json:"keystoreCertificateChain,omitempty"`
	KeystoreKey                   *string `json:"keystoreKey,omitempty"`
	DisableEndpointIdentification *bool   `json:"disableEndpointIdentification,omitempty"`
}

// KafkaLinkCreateParam represents the create payload for kafka linking.
type KafkaLinkCreateParam struct {
	LinkID          string                      `json:"linkId"`
	StartOffsetTime string                      `json:"startOffsetTime"`
	SourceCluster   KafkaLinkSourceClusterParam `json:"sourceCluster"`
}

// KafkaLinkVO describes a kafka link returned by the API.
type KafkaLinkVO struct {
	LinkID                 string                    `json:"linkId"`
	InstanceID             string                    `json:"instanceId"`
	StartOffsetTime        string                    `json:"startOffsetTime"`
	SourceCluster          *KafkaLinkSourceClusterVO `json:"sourceCluster,omitempty"`
	SourceSecurityProtocol *string                   `json:"sourceSecurityProtocol,omitempty"`
	SourceSaslMechanism    *string                   `json:"sourceSaslMechanism,omitempty"`
	SourceUser             *string                   `json:"sourceUser,omitempty"`
	GmtCreate              *time.Time                `json:"gmtCreate,omitempty"`
	GmtModified            *time.Time                `json:"gmtModified,omitempty"`
	Status                 *string                   `json:"status,omitempty"`
	ErrorMessage           *string                   `json:"errorMessage,omitempty"`
	Statistics             *KafkaLinkingStatisticVO  `json:"statistics,omitempty"`
}

// KafkaLinkSourceClusterVO describes the visible part of source cluster config.
type KafkaLinkSourceClusterVO struct {
	Endpoint         string  `json:"endpoint"`
	SecurityProtocol *string `json:"securityProtocol,omitempty"`
	SaslMechanism    *string `json:"saslMechanism,omitempty"`
	User             *string `json:"user,omitempty"`
}

// KafkaLinkingStatisticVO carries link level metrics.
type KafkaLinkingStatisticVO struct {
	LinkingThroughputIn  *int64 `json:"linkingThroughputIn,omitempty"`
	LinkingThroughputOut *int64 `json:"linkingThroughputOut,omitempty"`
	LinkingLag           *int64 `json:"linkingLag,omitempty"`
	LinkingLagTime       *int64 `json:"linkingLagTime,omitempty"`
}

// PageNumResultKafkaLinkVO models paginated link response.
type PageNumResultKafkaLinkVO struct {
	List []KafkaLinkVO `json:"list,omitempty"`
}

// KafkaLinkMirrorTopicParam identifies a source topic to mirror.
type KafkaLinkMirrorTopicParam struct {
	TopicName string `json:"topicName"`
}

// KafkaLinkMirrorTopicsCreateParam carries mirror topic create request.
type KafkaLinkMirrorTopicsCreateParam struct {
	SourceTopics []KafkaLinkMirrorTopicParam `json:"sourceTopics"`
}

// KafkaLinkMirrorTopicsUpdateParam updates mirror topic state.
type KafkaLinkMirrorTopicsUpdateParam struct {
	State string `json:"state"`
}

// MirrorTopicVO represents mirrored topic info.
type MirrorTopicVO struct {
	SourceTopicName    string                   `json:"sourceTopicName"`
	MirrorTopicName    *string                  `json:"mirrorTopicName,omitempty"`
	MirrorTopicID      *string                  `json:"mirrorTopicId,omitempty"`
	State              *KafkaLinkingStateVO     `json:"state,omitempty"`
	SubscribedGroupNum *int                     `json:"subscribedGroupNum,omitempty"`
	PromotedGroupNum   *int                     `json:"promotedGroupNum,omitempty"`
	Statistics         *KafkaLinkingStatisticVO `json:"statistics,omitempty"`
}

// MirrorTopicListVO wraps topics returned on creation.
type MirrorTopicListVO struct {
	Topics []MirrorTopicVO `json:"topics,omitempty"`
}

// KafkaLinkingStateVO describes state with optional error code.
type KafkaLinkingStateVO struct {
	State     *string `json:"state,omitempty"`
	ErrorCode *string `json:"errorCode,omitempty"`
}

// PageNumResultMirrorTopicVO models paginated mirror topics.
type PageNumResultMirrorTopicVO struct {
	List []MirrorTopicVO `json:"list,omitempty"`
}

// KafkaLinkMirrorGroupParam identifies a source consumer group to mirror.
type KafkaLinkMirrorGroupParam struct {
	ConsumerGroup string `json:"consumerGroup"`
}

// KafkaLinkMirrorGroupsCreateParam carries mirror group create request.
type KafkaLinkMirrorGroupsCreateParam struct {
	SourceGroups []KafkaLinkMirrorGroupParam `json:"sourceGroups"`
}

// MirrorConsumerGroupVO represents mirrored consumer group info.
type MirrorConsumerGroupVO struct {
	LinkID        *string              `json:"linkId,omitempty"`
	SourceGroupID string               `json:"sourceGroupId"`
	MirrorGroupID *string              `json:"mirrorGroupId,omitempty"`
	State         *KafkaLinkingStateVO `json:"state,omitempty"`
}

// MirrorConsumerGroupListVO wraps groups returned on creation.
type MirrorConsumerGroupListVO struct {
	Groups []MirrorConsumerGroupVO `json:"groups,omitempty"`
}

// PageNumResultMirrorConsumerGroupVO models paginated mirror consumer groups.
type PageNumResultMirrorConsumerGroupVO struct {
	List []MirrorConsumerGroupVO `json:"list,omitempty"`
}
