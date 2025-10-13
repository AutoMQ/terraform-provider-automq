package client

import "time"

// DeployProfileVO represents the response structure for a deployment profile
type DeployProfileVO struct {
	Name             *string                `json:"name,omitempty"`
	Provider         *string                `json:"provider,omitempty"`
	Region           *string                `json:"region,omitempty"`
	Vpc              *string                `json:"vpc,omitempty"`
	InstancePlatform *string                `json:"instancePlatform,omitempty"`
	CredentialType   *string                `json:"credentialType,omitempty"`
	ClusterId        *string                `json:"clusterId,omitempty"`
	GmtCreate        *time.Time             `json:"gmtCreate,omitempty"`
	GmtModified      *time.Time             `json:"gmtModified,omitempty"`
	Available        *bool                  `json:"available,omitempty"`
	System           *bool                  `json:"system,omitempty"`
	KubeConfig       *string                `json:"kubeConfig,omitempty"`
	OpsBucket        *BucketProfileDetailVO `json:"opsBucket,omitempty"`
	DnsZone          *string                `json:"dnsZone,omitempty"`
	InstanceProfile  *string                `json:"instanceProfile,omitempty"`
	CredentialId     *string                `json:"credentialId,omitempty"`
}

// BucketProfileVO
type PageNumResultBucketProfileVO struct {
	PageNum   *int32            `json:"pageNum,omitempty"`
	PageSize  *int32            `json:"pageSize,omitempty"`
	Total     *int64            `json:"total,omitempty"`
	List      []BucketProfileVO `json:"list,omitempty"`
	TotalPage *int64            `json:"totalPage,omitempty"`
}

// BucketProfileVO struct for BucketProfileVO
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

// BucketProfileDetailVO represents the response structure for a bucket profile
type BucketProfileDetailVO struct {
	Id          *string    `json:"id,omitempty"`
	BucketName  *string    `json:"bucketName,omitempty"`
	GmtCreate   *time.Time `json:"gmtCreate,omitempty"`
	GmtModified *time.Time `json:"gmtModified,omitempty"`
	Provider    *string    `json:"provider,omitempty"`
	Region      *string    `json:"region,omitempty"`
}
