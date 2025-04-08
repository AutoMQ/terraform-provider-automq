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
var _ datasource.DataSource = &DataBucketProfilesDataSource{}

func NewDataBucketProfilesDataSource() datasource.DataSource {
	return &DataBucketProfilesDataSource{}
}

// DeployProfileDataSource defines the data source implementation.
type DataBucketProfilesDataSource struct {
	client *client.Client
}

func (d *DataBucketProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_bucket_profiles"
}

func (d *DataBucketProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Using the `automq_data_bucket_profiles` data source, you can retrieve dataBucket profile information.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Target AutoMQ BYOC environment, this attribute is specified during the deployment and installation process.",
				Required:            true,
			},
			"profile_name": schema.StringAttribute{
				MarkdownDescription: "The name of the deployment profile.",
				Required:            true,
			},
			"data_buckets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the data bucket.",
							Computed:            true,
						},
						"bucket_name": schema.StringAttribute{
							MarkdownDescription: "The name of the data bucket.",
							Computed:            true,
						},
						"gmt_create": schema.StringAttribute{
							CustomType:          timetypes.RFC3339Type{},
							MarkdownDescription: "Creation time of the profile.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DataBucketProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataBucketProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.BucketProfilesModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, data.EnvironmentID.ValueString())

	// Get deploy profile from API
	profile, err := d.client.GetBucketProfiles(ctx, data.ProfileName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Deploy Profile",
			fmt.Sprintf("Unable to read deploy profile %s: %s", data.ProfileName.ValueString(), err),
		)
		return
	}
	models.FlattenBucketProfilesResource(profile, &data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
