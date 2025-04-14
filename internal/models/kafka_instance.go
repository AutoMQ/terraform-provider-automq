package models

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Instance states that represent the lifecycle of a Kafka instance
const (
	StateCreating = "Creating"
	StateRunning  = "Running"
	StateChanging = "Changing"
	StateDeleting = "Deleting"
	StateNotFound = "NotFound"
	StateError    = "Error"
	StateUnknown  = "Unknown"
)

type KafkaInstanceResourceModel struct {
	EnvironmentID  types.String       `tfsdk:"environment_id"`
	InstanceID     types.String       `tfsdk:"id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	DeployProfile  types.String       `tfsdk:"deploy_profile"`
	Version        types.String       `tfsdk:"version"`
	ComputeSpecs   *ComputeSpecsModel `tfsdk:"compute_specs"`
	Features       *FeaturesModel     `tfsdk:"features"`
	Endpoints      types.List         `tfsdk:"endpoints"`
	CreatedAt      timetypes.RFC3339  `tfsdk:"created_at"`
	LastUpdated    timetypes.RFC3339  `tfsdk:"last_updated"`
	InstanceStatus types.String       `tfsdk:"status"`
	Timeouts       timeouts.Value     `tfsdk:"timeouts"`
}

type KafkaInstanceModel struct {
	EnvironmentID  types.String          `tfsdk:"environment_id"`
	InstanceID     types.String          `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	Description    types.String          `tfsdk:"description"`
	DeployProfile  types.String          `tfsdk:"deploy_profile"`
	Version        types.String          `tfsdk:"version"`
	ComputeSpecs   *ComputeSpecsModel    `tfsdk:"compute_specs"`
	Features       *FeaturesSummaryModel `tfsdk:"features"`
	Endpoints      types.List            `tfsdk:"endpoints"`
	CreatedAt      timetypes.RFC3339     `tfsdk:"created_at"`
	LastUpdated    timetypes.RFC3339     `tfsdk:"last_updated"`
	InstanceStatus types.String          `tfsdk:"status"`
}

type InstanceAccessInfo struct {
	DisplayName      types.String `tfsdk:"display_name"`
	NetworkType      types.String `tfsdk:"network_type"`
	Protocol         types.String `tfsdk:"protocol"`
	Mechanisms       types.String `tfsdk:"mechanisms"`
	BootstrapServers types.String `tfsdk:"bootstrap_servers"`
}

type NetworkModel struct {
	Zone    types.String `tfsdk:"zone"`
	Subnets types.List   `tfsdk:"subnets"`
}

type ComputeSpecsModel struct {
	ReservedAku          types.Int64            `tfsdk:"reserved_aku"`
	Networks             []NetworkModel         `tfsdk:"networks"`
	KubernetesNodeGroups []NodeGroupModel       `tfsdk:"kubernetes_node_groups"`
	BucketProfiles       []BucketProfileIDModel `tfsdk:"bucket_profiles"`
}

type NodeGroupModel struct {
	ID types.String `tfsdk:"id"`
}

type FeaturesModel struct {
	WalMode         types.String   `tfsdk:"wal_mode"`
	InstanceConfigs types.Map      `tfsdk:"instance_configs"`
	Integrations    types.Set      `tfsdk:"integrations"`
	Security        *SecurityModel `tfsdk:"security"`
}

type FeaturesSummaryModel struct {
	WalMode         types.String          `tfsdk:"wal_mode"`
	InstanceConfigs types.Map             `tfsdk:"instance_configs"`
	Integrations    types.Set             `tfsdk:"integrations"`
	Security        *SecuritySummaryModel `tfsdk:"security"`
}

type BucketProfileIDModel struct {
	ID types.String `tfsdk:"id"`
}

type SecurityModel struct {
	AuthenticationMethods  types.Set    `tfsdk:"authentication_methods"`
	TransitEncryptionModes types.Set    `tfsdk:"transit_encryption_modes"`
	CertificateAuthority   types.String `tfsdk:"certificate_authority"`
	CertificateChain       types.String `tfsdk:"certificate_chain"`
	PrivateKey             types.String `tfsdk:"private_key"`
	DataEncryptionMode     types.String `tfsdk:"data_encryption_mode"`
}

type SecuritySummaryModel struct {
	AuthenticationMethods  types.Set    `tfsdk:"authentication_methods"`
	TransitEncryptionModes types.Set    `tfsdk:"transit_encryption_modes"`
	DataEncryptionMode     types.String `tfsdk:"data_encryption_mode"`
}

// ExpandKafkaInstanceResource converts a KafkaInstanceResourceModel to a client.InstanceCreateParam.
// It handles the conversion of all nested structures and validates required fields.
func ExpandKafkaInstanceResource(instance KafkaInstanceResourceModel, request *client.InstanceCreateParam) error {
	if request == nil {
		return fmt.Errorf("request parameter cannot be nil")
	}

	// Basic fields
	request.Name = instance.Name.ValueString()
	request.Description = instance.Description.ValueString()
	request.DeployProfile = instance.DeployProfile.ValueString()
	request.Version = instance.Version.ValueString()

	// Validate required fields
	if request.Name == "" {
		return fmt.Errorf("instance name is required")
	}

	if request.Version == "" {
		return fmt.Errorf("instance version is required")
	}

	// Compute Specs
	if instance.ComputeSpecs != nil {
		// Reserved AKU
		request.Spec = client.SpecificationParam{
			ReservedAku: int32(instance.ComputeSpecs.ReservedAku.ValueInt64()),
			NodeConfig:  &client.NodeConfigParam{},
		}

		// ignore Node Configs

		// Networks
		// Kubernetes Node Groups
		if len(instance.ComputeSpecs.KubernetesNodeGroups) > 0 {
			nodeGroups := make([]client.KubernetesNodeGroupParam, 0, len(instance.ComputeSpecs.KubernetesNodeGroups))
			for _, ng := range instance.ComputeSpecs.KubernetesNodeGroups {
				id := ng.ID.ValueString()
				nodeGroups = append(nodeGroups, client.KubernetesNodeGroupParam{
					Id: &id,
				})
			}
			request.Spec.KubernetesNodeGroups = nodeGroups
		} else if len(instance.ComputeSpecs.Networks) > 0 {
			networks := make([]client.InstanceNetworkParam, 0, len(instance.ComputeSpecs.Networks))
			for _, network := range instance.ComputeSpecs.Networks {
				subnet := ""
				if !network.Subnets.IsNull() {
					subnets := ExpandStringValueList(network.Subnets)
					if len(subnets) > 0 {
						subnet = subnets[0]
					}
				}
				networks = append(networks, client.InstanceNetworkParam{
					Zone:   network.Zone.ValueString(),
					Subnet: &subnet,
				})
			}
			request.Spec.Networks = networks
		}
		// Bucket Profiles
		if len(instance.ComputeSpecs.BucketProfiles) > 0 {
			bucketProfiles := make([]client.BucketProfileBindParam, 0, len(instance.ComputeSpecs.BucketProfiles))
			for _, ng := range instance.ComputeSpecs.BucketProfiles {
				id := ng.ID.ValueString()
				bucketProfiles = append(bucketProfiles, client.BucketProfileBindParam{
					Id: &id,
				})
			}
			request.Spec.BucketProfiles = bucketProfiles
		}
	}

	// Features
	if instance.Features != nil {
		// WAL
		walMode := instance.Features.WalMode.ValueString()
		request.Features = &client.InstanceFeatureParam{
			WalMode: &walMode,
		}

		// Security
		if instance.Features.Security != nil {
			if !instance.Features.Security.DataEncryptionMode.IsNull() {
				dataEncryptionMode := instance.Features.Security.DataEncryptionMode.ValueString()
				request.Features.Security = &client.InstanceSecurityParam{
					DataEncryptionMode: &dataEncryptionMode,
				}
			}

			if !instance.Features.Security.AuthenticationMethods.IsNull() {
				var authMethods []string
				diags := instance.Features.Security.AuthenticationMethods.ElementsAs(context.TODO(), &authMethods, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse authentication methods: %v", diags)
				}
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.AuthenticationMethods = authMethods
			}

			if !instance.Features.Security.TransitEncryptionModes.IsNull() {
				var encryptionModes []string
				diags := instance.Features.Security.TransitEncryptionModes.ElementsAs(context.TODO(), &encryptionModes, false)
				if diags.HasError() {
					return fmt.Errorf("failed to parse transit encryption modes: %v", diags)
				}
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.TransitEncryptionModes = encryptionModes
			}

			if !instance.Features.Security.CertificateAuthority.IsNull() {
				certAuth := instance.Features.Security.CertificateAuthority.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.CertificateAuthority = &certAuth
			}

			if !instance.Features.Security.CertificateChain.IsNull() {
				certChain := instance.Features.Security.CertificateChain.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.CertificateChain = &certChain
			}

			if !instance.Features.Security.PrivateKey.IsNull() {
				privateKey := instance.Features.Security.PrivateKey.ValueString()
				if request.Features.Security == nil {
					request.Features.Security = &client.InstanceSecurityParam{}
				}
				request.Features.Security.PrivateKey = &privateKey
			}
		}

		// Integrations
		if !instance.Features.Integrations.IsNull() && len(instance.Features.Integrations.Elements()) > 0 {
			ids := ExpandSetValueList(instance.Features.Integrations)
			integrations := make([]client.IntegrationBindParam, 0, len(ids))
			for _, id := range ids {
				integrationID := id
				integrations = append(integrations, client.IntegrationBindParam{
					Id: &integrationID,
				})
			}
			request.Features.Integrations = integrations
		}

		// Instance Configs
		if !instance.Features.InstanceConfigs.IsNull() {
			instanceConfigs := ExpandStringValueMap(instance.Features.InstanceConfigs)
			request.Features.InstanceConfigs = instanceConfigs
		}
	}
	return nil
}

// ConvertKafkaInstanceModel copies data from a KafkaInstanceResourceModel to a KafkaInstanceModel.
// This function is used when converting between different model representations.
func ConvertKafkaInstanceModel(resource *KafkaInstanceResourceModel, model *KafkaInstanceModel) error {
	if resource == nil || model == nil {
		return fmt.Errorf("both resource and model parameters must not be nil")
	}

	model.EnvironmentID = resource.EnvironmentID
	model.InstanceID = resource.InstanceID
	model.Name = resource.Name
	model.Description = resource.Description
	model.DeployProfile = resource.DeployProfile
	model.Version = resource.Version
	model.ComputeSpecs = resource.ComputeSpecs
	model.Features = &FeaturesSummaryModel{
		WalMode:         resource.Features.WalMode,
		InstanceConfigs: resource.Features.InstanceConfigs,
		Integrations:    resource.Features.Integrations,
		Security: &SecuritySummaryModel{
			AuthenticationMethods:  resource.Features.Security.AuthenticationMethods,
			TransitEncryptionModes: resource.Features.Security.TransitEncryptionModes,
			DataEncryptionMode:     resource.Features.Security.DataEncryptionMode,
		},
	}
	model.Endpoints = resource.Endpoints
	model.CreatedAt = resource.CreatedAt
	model.LastUpdated = resource.LastUpdated
	model.InstanceStatus = resource.InstanceStatus
	return nil
}

// FlattenKafkaInstanceBasicModel converts a client.InstanceSummaryVO to a KafkaInstanceResourceModel.
// It handles the basic fields of the instance model.
func FlattenKafkaInstanceBasicModel(instance *client.InstanceSummaryVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Instance",
			"Cannot flatten nil instance",
		)}
	}

	if resource == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Resource",
			"Cannot flatten instance into nil resource",
		)}
	}

	// Basic fields
	if instance.InstanceId != nil {
		resource.InstanceID = types.StringValue(*instance.InstanceId)
	}
	if instance.Name != nil {
		resource.Name = types.StringValue(*instance.Name)
	}
	if instance.Description != nil {
		resource.Description = types.StringValue(*instance.Description)
	}
	if instance.DeployProfile != nil {
		resource.DeployProfile = types.StringValue(*instance.DeployProfile)
	}
	if instance.Version != nil {
		resource.Version = types.StringValue(*instance.Version)
	}
	if instance.State != nil {
		resource.InstanceStatus = types.StringValue(*instance.State)
	}

	// Timestamps
	if instance.GmtCreate != nil {
		resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(instance.GmtCreate)
	}
	if instance.GmtModified != nil {
		resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(instance.GmtModified)
	}

	return diags
}

// FlattenKafkaInstanceModel converts a client.InstanceVO to a KafkaInstanceResourceModel.
// It handles all fields including nested structures.
func FlattenKafkaInstanceModel(instance *client.InstanceVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if instance == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Instance",
			"Cannot flatten nil instance",
		)}
	}

	if resource == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Resource",
			"Cannot flatten instance into nil resource",
		)}
	}

	// Basic fields
	if instance.InstanceId != nil {
		resource.InstanceID = types.StringValue(*instance.InstanceId)
	}
	if instance.Name != nil {
		resource.Name = types.StringValue(*instance.Name)
	}
	if instance.Description != nil {
		resource.Description = types.StringValue(*instance.Description)
	}
	if instance.DeployProfile != nil {
		resource.DeployProfile = types.StringValue(*instance.DeployProfile)
	}
	if instance.Version != nil {
		resource.Version = types.StringValue(*instance.Version)
	}
	if instance.State != nil {
		resource.InstanceStatus = types.StringValue(*instance.State)
	}

	// Compute Specs
	if instance.Spec != nil {
		if resource.ComputeSpecs == nil {
			resource.ComputeSpecs = &ComputeSpecsModel{
				ReservedAku:          types.Int64Null(),
				Networks:             []NetworkModel{},
				KubernetesNodeGroups: []NodeGroupModel{},
				BucketProfiles:       []BucketProfileIDModel{},
			}
		}
		// Reserved AKU
		if instance.Spec.ReservedAku != nil {
			resource.ComputeSpecs.ReservedAku = types.Int64Value(int64(*instance.Spec.ReservedAku))
		}

		// Kubernetes Node Groups
		if instance.Spec.KubernetesNodeGroups != nil {
			nodeGroups := make([]NodeGroupModel, 0, len(instance.Spec.KubernetesNodeGroups))
			for _, ng := range instance.Spec.KubernetesNodeGroups {
				if ng.Id != nil {
					nodeGroups = append(nodeGroups, NodeGroupModel{
						ID: types.StringValue(*ng.Id),
					})
				}
			}
			resource.ComputeSpecs.KubernetesNodeGroups = nodeGroups
			// Networks
		} else if instance.Spec.Networks != nil {
			networks, networkDiags := flattenNetworks(instance.Spec.Networks)
			if networkDiags.HasError() {
				diags.Append(networkDiags...)
			} else {
				resource.ComputeSpecs.Networks = networks
			}
		}

		// Bucket Profiles
		if instance.Spec.BucketProfiles != nil {
			bucketProfiles := make([]BucketProfileIDModel, 0, len(instance.Spec.BucketProfiles))
			for _, bp := range instance.Spec.BucketProfiles {
				if bp.Id != nil {
					bucketProfiles = append(bucketProfiles, BucketProfileIDModel{
						ID: types.StringValue(*bp.Id),
					})
				}
			}
			resource.ComputeSpecs.BucketProfiles = bucketProfiles
		}
	}

	// Features
	if instance.Features != nil {
		if resource.Features == nil {
			resource.Features = &FeaturesModel{}
		}

		// WAL Mode
		if instance.Features.WalMode != nil {
			resource.Features.WalMode = types.StringValue(*instance.Features.WalMode)
		}

		// Security
		if instance.Features.Security != nil {
			if resource.Features.Security == nil {
				resource.Features.Security = &SecurityModel{}
			}

			if instance.Features.Security.DataEncryptionMode != nil {
				resource.Features.Security.DataEncryptionMode = types.StringValue(*instance.Features.Security.DataEncryptionMode)
			}

			// Authentication Methods
			if instance.Features.Security.AuthenticationMethods != nil {
				values := make([]attr.Value, len(instance.Features.Security.AuthenticationMethods))
				for i, v := range instance.Features.Security.AuthenticationMethods {
					values[i] = types.StringValue(v)
				}
				set, listDiags := types.SetValue(types.StringType, values)
				if !listDiags.HasError() {
					resource.Features.Security.AuthenticationMethods = set
				} else {
					diags.Append(listDiags...)
				}
			}
			// Transit Encryption Modes
			if instance.Features.Security.TransitEncryptionModes != nil {
				values := make([]attr.Value, len(instance.Features.Security.TransitEncryptionModes))
				for i, v := range instance.Features.Security.TransitEncryptionModes {
					values[i] = types.StringValue(v)
				}
				set, listDiags := types.SetValue(types.StringType, values)
				if !listDiags.HasError() {
					resource.Features.Security.TransitEncryptionModes = set
				} else {
					diags.Append(listDiags...)
				}
			}
		}
	}

	// Timestamps
	if instance.GmtCreate != nil {
		resource.CreatedAt = timetypes.NewRFC3339TimePointerValue(instance.GmtCreate)
	}
	if instance.GmtModified != nil {
		resource.LastUpdated = timetypes.NewRFC3339TimePointerValue(instance.GmtModified)
	}

	return diags
}

// FlattenKafkaInstanceModelWithIntegrations adds integration information to the KafkaInstanceResourceModel.
func FlattenKafkaInstanceModelWithIntegrations(integrations []client.IntegrationVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if integrations == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Integrations",
			"Cannot flatten nil integrations",
		)}
	}

	if resource.Features == nil {
		resource.Features = &FeaturesModel{}
	}
	// Handle integrations if present
	if len(integrations) > 0 {
		integrationIds := make([]attr.Value, 0, len(integrations))
		for _, integration := range integrations {
			integrationIds = append(integrationIds, types.StringValue(integration.Code))
		}
		resource.Features.Integrations = types.SetValueMust(types.StringType, integrationIds)
	} else if len(integrations) == 0 {
		resource.Features.Integrations = types.SetNull(types.StringType)
	}
	return diags
}

// FlattenKafkaInstanceModelWithEndpoints adds endpoint information to the KafkaInstanceResourceModel.
func FlattenKafkaInstanceModelWithEndpoints(endpoints []client.InstanceAccessInfoVO, resource *KafkaInstanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if endpoints == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Endpoints",
			"Cannot flatten nil endpoints",
		)}
	}

	// Handle endpoints
	if len(endpoints) > 0 {
		if err := populateInstanceAccessInfoList(context.Background(), resource, endpoints); err.HasError() {
			diags.Append(err...)
		}
	}

	return diags
}

// populateInstanceAccessInfoList converts endpoint information into a list of InstanceAccessInfo.
func populateInstanceAccessInfoList(ctx context.Context, data *KafkaInstanceResourceModel, in []client.InstanceAccessInfoVO) diag.Diagnostics {
	var diags diag.Diagnostics

	if data == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Resource",
			"Cannot populate access info for nil resource",
		)}
	}

	instanceAccessInfoList := make([]InstanceAccessInfo, 0, len(in))
	for _, item := range in {
		if item.DisplayName == nil || item.NetworkType == nil || item.Protocol == nil ||
			item.Mechanisms == nil || item.BootstrapServers == nil {
			continue
		}

		instanceAccessInfoList = append(instanceAccessInfoList, InstanceAccessInfo{
			DisplayName:      types.StringValue(*item.DisplayName),
			NetworkType:      types.StringValue(*item.NetworkType),
			Protocol:         types.StringValue(*item.Protocol),
			Mechanisms:       types.StringValue(*item.Mechanisms),
			BootstrapServers: types.StringValue(*item.BootstrapServers),
		})
	}

	endpointsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
		"display_name":      types.StringType,
		"network_type":      types.StringType,
		"protocol":          types.StringType,
		"mechanisms":        types.StringType,
		"bootstrap_servers": types.StringType,
	}}, instanceAccessInfoList)

	if !diags.HasError() {
		data.Endpoints = endpointsList
	}

	return diags
}

// flattenNetworks converts network information into a slice of NetworkModel.
func flattenNetworks(networks []client.InstanceZoneNetworkVO) ([]NetworkModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if networks == nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid Networks",
			"Cannot flatten nil networks",
		)}
	}

	networksModel := make([]NetworkModel, 0, len(networks))
	for _, network := range networks {
		if network.Zone == nil {
			continue
		}

		zone := types.StringValue(*network.Zone)
		subnets := make([]attr.Value, 0)
		if network.Subnet != nil {
			subnets = append(subnets, types.StringValue(*network.Subnet))
		}

		subnetList, listDiags := types.ListValue(types.StringType, subnets)
		if listDiags.HasError() {
			diags.Append(listDiags...)
			continue
		}

		networksModel = append(networksModel, NetworkModel{
			Zone:    zone,
			Subnets: subnetList,
		})
	}

	return networksModel, diags
}
