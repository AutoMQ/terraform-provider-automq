package models

import (
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeployProfileModel describes the data source data model.
type DeployProfileModel struct {
	EnvironmentID    types.String              `tfsdk:"environment_id"`
	Name             types.String              `tfsdk:"name"`
	CloudProvider    types.String              `tfsdk:"cloud_provider"`
	Region           types.String              `tfsdk:"region"`
	Vpc              types.String              `tfsdk:"vpc"`
	InstancePlatform types.String              `tfsdk:"instance_platform"`
	GmtCreate        timetypes.RFC3339         `tfsdk:"gmt_create"`
	GmtModified      timetypes.RFC3339         `tfsdk:"gmt_modified"`
	Available        types.Bool                `tfsdk:"available"`
	System           types.Bool                `tfsdk:"system"`
	OpsBucket        *BucketProfileDetailModel `tfsdk:"ops_bucket"`
	DnsZone          types.String              `tfsdk:"dns_zone"`
	InstanceProfile  types.String              `tfsdk:"instance_profile"`
}

// BucketProfileModel describes the bucket profile data model.
type BucketProfileModel struct {
	ID         types.String      `tfsdk:"id"`
	BucketName types.String      `tfsdk:"bucket_name"`
	GmtCreate  timetypes.RFC3339 `tfsdk:"gmt_create"`
}

type BucketProfilesModel struct {
	EnvironmentID  types.String         `tfsdk:"environment_id"`
	ProfileName    types.String         `tfsdk:"profile_name"`
	BucketProfiles []BucketProfileModel `tfsdk:"data_buckets"`
}

type BucketProfileDetailModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Provider   types.String `tfsdk:"provider"`
	Region     types.String `tfsdk:"region"`
}

func FlattenDeployProfileResource(profile *client.DeployProfileVO, data *DeployProfileModel) {
	// Map response body to model
	data.CloudProvider = types.StringValue(*profile.Provider)
	data.Region = types.StringValue(*profile.Region)
	data.Vpc = types.StringValue(*profile.Vpc)
	data.InstancePlatform = types.StringValue(*profile.InstancePlatform)
	data.GmtCreate = timetypes.NewRFC3339TimePointerValue(profile.GmtCreate)
	data.GmtModified = timetypes.NewRFC3339TimePointerValue(profile.GmtModified)
	data.Available = types.BoolValue(*profile.Available)
	data.System = types.BoolValue(*profile.System)
	data.DnsZone = types.StringValue(*profile.DnsZone)
	data.InstanceProfile = types.StringValue(*profile.InstanceProfile)

	// Map ops bucket
	if profile.OpsBucket != nil {
		data.OpsBucket = &BucketProfileDetailModel{
			BucketName: types.StringValue(*profile.OpsBucket.BucketName),
			Provider:   types.StringValue(*profile.OpsBucket.Provider),
			Region:     types.StringValue(*profile.OpsBucket.Region),
		}
	}

	// Map data buckets
	// if profile.DataBuckets != nil {
	// 	data.DataBuckets = make([]BucketProfileDetailModel, 0, len(profile.DataBuckets))
	// 	for _, bucket := range profile.DataBuckets {
	// 		bucketModel := BucketProfileDetailModel{
	// 			ID:          types.StringValue(*bucket.Id),
	// 			BucketName:  types.StringValue(*bucket.BucketName),
	// 			GmtCreate:   timetypes.NewRFC3339TimePointerValue(bucket.GmtCreate),
	// 			GmtModified: timetypes.NewRFC3339TimePointerValue(bucket.GmtModified),
	// 			Provider:    types.StringValue(*bucket.Provider),
	// 			Region:      types.StringValue(*bucket.Region),
	// 		}
	// 		data.DataBuckets = append(data.DataBuckets, bucketModel)
	// 	}
	// }
}

func FlattenBucketProfilesResource(bucket *client.PageNumResultBucketProfileVO, data *BucketProfilesModel) {
	// Map response body to model
	data.BucketProfiles = make([]BucketProfileModel, 0, len(bucket.List))
	for _, bucket := range bucket.List {
		bucketModel := BucketProfileModel{
			ID:         types.StringValue(*bucket.Id),
			BucketName: types.StringValue(*bucket.BucketName),
			GmtCreate:  timetypes.NewRFC3339TimePointerValue(bucket.GmtCreate),
		}
		data.BucketProfiles = append(data.BucketProfiles, bucketModel)
	}
}
