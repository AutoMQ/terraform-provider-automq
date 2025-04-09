package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DeployProfileDataSource{}

func NewDeployProfileDataSource() datasource.DataSource {
	return &DeployProfileDataSource{}
}

// DeployProfileDataSource defines the data source implementation.
type DeployProfileDataSource struct {
	client *client.Client
}

func (d *DeployProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deploy_profile"
}

func (d *DeployProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)<br><br>Using the `automq_deploy_profile` data source, you can retrieve deployment profile information.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the deployment profile.",
				Required:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "The cloud provider (e.g., aws).",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The region where the profile is deployed.",
				Computed:            true,
			},
			"vpc": schema.StringAttribute{
				MarkdownDescription: "The VPC ID.",
				Computed:            true,
			},
			"instance_platform": schema.StringAttribute{
				MarkdownDescription: "The instance platform type.",
				Computed:            true,
			},
			"gmt_create": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "Creation time of the profile.",
				Computed:            true,
			},
			"gmt_modified": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "Last modification time of the profile.",
				Computed:            true,
			},
			"available": schema.BoolAttribute{
				MarkdownDescription: "Whether the profile is available.",
				Computed:            true,
			},
			"system": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a system profile.",
				Computed:            true,
			},
			"ops_bucket": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"bucket_name": schema.StringAttribute{
						MarkdownDescription: "The name of the operations bucket.",
						Computed:            true,
					},
					"provider": schema.StringAttribute{
						MarkdownDescription: "The cloud provider of the operations bucket.",
						Computed:            true,
					},
					"region": schema.StringAttribute{
						MarkdownDescription: "The region of the operations bucket.",
						Computed:            true,
					},
				},
			},
			"dns_zone": schema.StringAttribute{
				MarkdownDescription: "The DNS zone ID.",
				Computed:            true,
			},
			"instance_profile": schema.StringAttribute{
				MarkdownDescription: "The instance profile ARN.",
				Computed:            true,
			},
		},
	}
}

func (d *DeployProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *DeployProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.DeployProfileModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, data.EnvironmentID.ValueString())

	// Get deploy profile from API
	profile, err := d.client.GetDeployProfile(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Deploy Profile",
			fmt.Sprintf("Unable to read deploy profile %s: %s", data.Name.ValueString(), err),
		)
		return
	}
	models.FlattenDeployProfileResource(profile, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
