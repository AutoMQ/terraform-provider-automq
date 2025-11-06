package models

import (
	"strings"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KafkaLinkingResourceModel represents the Terraform state for automq_kafka_linking.
type KafkaLinkingResourceModel struct {
	EnvironmentID   types.String                 `tfsdk:"environment_id"`
	InstanceID      types.String                 `tfsdk:"instance_id"`
	LinkID          types.String                 `tfsdk:"link_id"`
	StartOffsetTime types.String                 `tfsdk:"start_offset_time"`
	SourceCluster   *KafkaLinkSourceClusterModel `tfsdk:"source_cluster"`
	Status          types.String                 `tfsdk:"status"`
	CreatedAt       timetypes.RFC3339            `tfsdk:"created_at"`
	LastUpdated     timetypes.RFC3339            `tfsdk:"last_updated"`
	ErrorMessage    types.String                 `tfsdk:"error_message"`
}

// KafkaLinkSourceClusterModel mirrors the nested source_cluster block.
type KafkaLinkSourceClusterModel struct {
	Endpoint                      types.String `tfsdk:"endpoint"`
	SecurityProtocol              types.String `tfsdk:"security_protocol"`
	SaslMechanism                 types.String `tfsdk:"sasl_mechanism"`
	User                          types.String `tfsdk:"user"`
	Password                      types.String `tfsdk:"password"`
	TruststoreCertificates        types.String `tfsdk:"truststore_certificates"`
	KeystoreCertificateChain      types.String `tfsdk:"keystore_certificate_chain"`
	KeystoreKey                   types.String `tfsdk:"keystore_key"`
	DisableEndpointIdentification types.Bool   `tfsdk:"disable_endpoint_identification"`
}

// KafkaMirrorTopicResourceModel represents automq_kafka_mirror_topic schema.
type KafkaMirrorTopicResourceModel struct {
	EnvironmentID   types.String `tfsdk:"environment_id"`
	InstanceID      types.String `tfsdk:"instance_id"`
	LinkID          types.String `tfsdk:"link_id"`
	SourceTopicName types.String `tfsdk:"source_topic_name"`
	State           types.String `tfsdk:"state"`
	MirrorTopicID   types.String `tfsdk:"mirror_topic_id"`
	MirrorTopicName types.String `tfsdk:"mirror_topic_name"`
	ErrorCode       types.String `tfsdk:"error_code"`
}

// KafkaMirrorGroupResourceModel represents automq_kafka_mirror_group schema.
type KafkaMirrorGroupResourceModel struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	InstanceID    types.String `tfsdk:"instance_id"`
	LinkID        types.String `tfsdk:"link_id"`
	SourceGroupID types.String `tfsdk:"source_group_id"`
	MirrorGroupID types.String `tfsdk:"mirror_group_id"`
	State         types.String `tfsdk:"state"`
	ErrorCode     types.String `tfsdk:"error_code"`
}

// ExpandKafkaLinkCreateParam builds API payload from Terraform model.
func ExpandKafkaLinkCreateParam(model *KafkaLinkingResourceModel) (client.KafkaLinkCreateParam, diag.Diagnostics) {
	var diags diag.Diagnostics
	param := client.KafkaLinkCreateParam{
		LinkID:          model.LinkID.ValueString(),
		StartOffsetTime: model.StartOffsetTime.ValueString(),
	}

	if model.SourceCluster == nil {
		diags.AddError("Missing source cluster configuration", "source_cluster block must be provided")
		return param, diags
	}

	param.SourceCluster = expandKafkaLinkSourceCluster(model.SourceCluster)
	return param, diags
}

func expandKafkaLinkSourceCluster(model *KafkaLinkSourceClusterModel) client.KafkaLinkSourceClusterParam {
	var securityProtocol *string
	if !model.SecurityProtocol.IsNull() && !model.SecurityProtocol.IsUnknown() {
		value := model.SecurityProtocol.ValueString()
		securityProtocol = &value
	}
	var saslMechanism *string
	if !model.SaslMechanism.IsNull() && !model.SaslMechanism.IsUnknown() {
		value := model.SaslMechanism.ValueString()
		saslMechanism = &value
	}
	var user *string
	if !model.User.IsNull() && !model.User.IsUnknown() {
		value := model.User.ValueString()
		user = &value
	}
	var password *string
	if !model.Password.IsNull() && !model.Password.IsUnknown() {
		value := model.Password.ValueString()
		password = &value
	}
	var truststore *string
	if !model.TruststoreCertificates.IsNull() && !model.TruststoreCertificates.IsUnknown() {
		value := model.TruststoreCertificates.ValueString()
		truststore = &value
	}
	var keystoreChain *string
	if !model.KeystoreCertificateChain.IsNull() && !model.KeystoreCertificateChain.IsUnknown() {
		value := model.KeystoreCertificateChain.ValueString()
		keystoreChain = &value
	}
	var keystoreKey *string
	if !model.KeystoreKey.IsNull() && !model.KeystoreKey.IsUnknown() {
		value := model.KeystoreKey.ValueString()
		keystoreKey = &value
	}
	var disableEndpointIdentification *bool
	if !model.DisableEndpointIdentification.IsNull() && !model.DisableEndpointIdentification.IsUnknown() {
		value := model.DisableEndpointIdentification.ValueBool()
		disableEndpointIdentification = &value
	}

	return client.KafkaLinkSourceClusterParam{
		Endpoint:                      model.Endpoint.ValueString(),
		SecurityProtocol:              securityProtocol,
		SaslMechanism:                 saslMechanism,
		User:                          user,
		Password:                      password,
		TruststoreCertificates:        truststore,
		KeystoreCertificateChain:      keystoreChain,
		KeystoreKey:                   keystoreKey,
		DisableEndpointIdentification: disableEndpointIdentification,
	}
}

// FlattenKafkaLink populates Terraform state from API response, preserving sensitive values when absent.
func FlattenKafkaLink(link *client.KafkaLinkVO, state *KafkaLinkingResourceModel, prior *KafkaLinkingResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if link == nil {
		return diags
	}

	if prior != nil {
		state.EnvironmentID = prior.EnvironmentID
	} else {
		state.EnvironmentID = types.StringNull()
	}

	state.InstanceID = types.StringValue(link.InstanceID)
	state.LinkID = types.StringValue(link.LinkID)
	state.StartOffsetTime = types.StringValue(link.StartOffsetTime)

	if link.Status != nil {
		state.Status = types.StringValue(*link.Status)
	} else if prior != nil {
		state.Status = prior.Status
	} else {
		state.Status = types.StringNull()
	}

	if link.ErrorMessage != nil && *link.ErrorMessage != "" {
		state.ErrorMessage = types.StringValue(*link.ErrorMessage)
	} else {
		state.ErrorMessage = types.StringNull()
		if prior != nil {
			state.ErrorMessage = prior.ErrorMessage
		}
	}

	if link.GmtCreate != nil {
		state.CreatedAt = timetypes.NewRFC3339TimePointerValue(link.GmtCreate)
	} else if prior != nil {
		state.CreatedAt = prior.CreatedAt
	} else {
		state.CreatedAt = timetypes.RFC3339{}
	}

	if link.GmtModified != nil {
		state.LastUpdated = timetypes.NewRFC3339TimePointerValue(link.GmtModified)
	} else if prior != nil {
		state.LastUpdated = prior.LastUpdated
	} else {
		state.LastUpdated = timetypes.RFC3339{}
	}

	state.SourceCluster = flattenKafkaLinkSourceCluster(link, prior)

	return diags
}

func flattenKafkaLinkSourceCluster(link *client.KafkaLinkVO, prior *KafkaLinkingResourceModel) *KafkaLinkSourceClusterModel {
	var previous *KafkaLinkSourceClusterModel
	if prior != nil {
		previous = prior.SourceCluster
	}

	result := &KafkaLinkSourceClusterModel{}
	if previous != nil {
		*result = *previous
	} else {
		result.Endpoint = types.StringNull()
		result.SecurityProtocol = types.StringNull()
		result.SaslMechanism = types.StringNull()
		result.User = types.StringNull()
		result.Password = types.StringNull()
		result.TruststoreCertificates = types.StringNull()
		result.KeystoreCertificateChain = types.StringNull()
		result.KeystoreKey = types.StringNull()
		result.DisableEndpointIdentification = types.BoolNull()
	}

	if link.SourceCluster != nil {
		result.Endpoint = types.StringValue(link.SourceCluster.Endpoint)
		if link.SourceCluster.SecurityProtocol != nil {
			result.SecurityProtocol = types.StringValue(*link.SourceCluster.SecurityProtocol)
		}
		if link.SourceCluster.SaslMechanism != nil {
			result.SaslMechanism = types.StringValue(*link.SourceCluster.SaslMechanism)
		}
		if link.SourceCluster.User != nil {
			result.User = types.StringValue(*link.SourceCluster.User)
		}
	}

	if link.SourceSecurityProtocol != nil {
		result.SecurityProtocol = types.StringValue(*link.SourceSecurityProtocol)
	}
	if link.SourceSaslMechanism != nil {
		result.SaslMechanism = types.StringValue(*link.SourceSaslMechanism)
	}
	if link.SourceUser != nil {
		result.User = types.StringValue(*link.SourceUser)
	}

	if previous != nil {
		result.Password = previous.Password
		result.TruststoreCertificates = previous.TruststoreCertificates
		result.KeystoreCertificateChain = previous.KeystoreCertificateChain
		result.KeystoreKey = previous.KeystoreKey
		result.DisableEndpointIdentification = previous.DisableEndpointIdentification
	}

	return result
}

// FlattenMirrorTopic maps API topic data to Terraform state.
func FlattenMirrorTopic(topic *client.MirrorTopicVO, state *KafkaMirrorTopicResourceModel, prior *KafkaMirrorTopicResourceModel) {
	if topic == nil {
		return
	}
	if prior != nil {
		state.EnvironmentID = prior.EnvironmentID
		state.InstanceID = prior.InstanceID
		state.LinkID = prior.LinkID
	} else {
		state.EnvironmentID = types.StringNull()
		state.InstanceID = types.StringNull()
		state.LinkID = types.StringNull()
	}
	state.SourceTopicName = types.StringValue(topic.SourceTopicName)
	if topic.MirrorTopicID != nil {
		state.MirrorTopicID = types.StringValue(*topic.MirrorTopicID)
	} else if prior != nil {
		state.MirrorTopicID = prior.MirrorTopicID
	} else {
		state.MirrorTopicID = types.StringNull()
	}
	if topic.MirrorTopicName != nil {
		state.MirrorTopicName = types.StringValue(*topic.MirrorTopicName)
	} else if prior != nil {
		state.MirrorTopicName = prior.MirrorTopicName
	} else {
		state.MirrorTopicName = types.StringNull()
	}
	if topic.State != nil && topic.State.State != nil {
		state.State = types.StringValue(*topic.State.State)
		if topic.State.ErrorCode != nil {
			state.ErrorCode = types.StringValue(*topic.State.ErrorCode)
		} else {
			state.ErrorCode = types.StringNull()
		}
	} else {
		if prior != nil {
			state.State = prior.State
			state.ErrorCode = prior.ErrorCode
		} else {
			state.State = types.StringNull()
			state.ErrorCode = types.StringNull()
		}
	}
}

// FlattenMirrorConsumerGroup maps API group data to Terraform state.
func FlattenMirrorConsumerGroup(group *client.MirrorConsumerGroupVO, state *KafkaMirrorGroupResourceModel, prior *KafkaMirrorGroupResourceModel) {
	if group == nil {
		return
	}
	if prior != nil {
		state.EnvironmentID = prior.EnvironmentID
		state.InstanceID = prior.InstanceID
		state.LinkID = prior.LinkID
	} else {
		state.EnvironmentID = types.StringNull()
		state.InstanceID = types.StringNull()
		state.LinkID = types.StringNull()
	}
	state.SourceGroupID = types.StringValue(group.SourceGroupID)
	if group.MirrorGroupID != nil {
		state.MirrorGroupID = types.StringValue(*group.MirrorGroupID)
	} else if prior != nil {
		state.MirrorGroupID = prior.MirrorGroupID
	} else {
		state.MirrorGroupID = types.StringNull()
	}
	if group.State != nil && group.State.State != nil {
		state.State = types.StringValue(*group.State.State)
		if group.State.ErrorCode != nil {
			state.ErrorCode = types.StringValue(*group.State.ErrorCode)
		} else {
			state.ErrorCode = types.StringNull()
		}
	} else {
		if prior != nil {
			state.State = prior.State
			state.ErrorCode = prior.ErrorCode
		} else {
			state.State = types.StringNull()
			state.ErrorCode = types.StringNull()
		}
	}
}

// BuildMirrorTopicCreateParam converts Terraform model into API create payload.
func BuildMirrorTopicCreateParam(model *KafkaMirrorTopicResourceModel) client.KafkaLinkMirrorTopicsCreateParam {
	return client.KafkaLinkMirrorTopicsCreateParam{
		SourceTopics: []client.KafkaLinkMirrorTopicParam{
			{
				TopicName: model.SourceTopicName.ValueString(),
			},
		},
	}
}

// BuildMirrorTopicUpdateParam converts Terraform model into API update payload.
func BuildMirrorTopicUpdateParam(model *KafkaMirrorTopicResourceModel) client.KafkaLinkMirrorTopicsUpdateParam {
	state := strings.ToUpper(model.State.ValueString())
	return client.KafkaLinkMirrorTopicsUpdateParam{
		State: state,
	}
}

// BuildMirrorGroupCreateParam converts Terraform model into API create payload.
func BuildMirrorGroupCreateParam(model *KafkaMirrorGroupResourceModel) client.KafkaLinkMirrorGroupsCreateParam {
	return client.KafkaLinkMirrorGroupsCreateParam{
		SourceGroups: []client.KafkaLinkMirrorGroupParam{
			{
				ConsumerGroup: model.SourceGroupID.ValueString(),
			},
		},
	}
}
