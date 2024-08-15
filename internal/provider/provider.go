package provider

import (
	"context"
	"os"
	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure AutoMQProvider satisfies various provider interfaces.
var _ provider.Provider = &AutoMQProvider{}
var _ provider.ProviderWithFunctions = &AutoMQProvider{}

// AutoMQProvider defines the provider implementation.
type AutoMQProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// autoMQProviderModel describes the provider data model.
type autoMQProviderModel struct {
	BYOCAccessKey types.String `tfsdk:"automq_byoc_access_key_id"`
	BYOCSecretKey types.String `tfsdk:"automq_byoc_secret_key"`
	BYOCHost      types.String `tfsdk:"automq_byoc_host"`
}

func (p *AutoMQProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "automq"
	resp.Version = p.version
}

func (p *AutoMQProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"automq_byoc_access_key_id": schema.StringAttribute{
				MarkdownDescription: "Set the Access Key Id of client. AutoMQ Cloud (BYOC) requires Access Keys to manage access and authentication to different parts of the service. An Access Key consists of an access key id and a secret key. You can create and manage Access Keys by using the AutoMQ Cloud BYOC Console. Learn more about AutoMQ Cloud BYOC Console access [here](https://docs.automq.com/automq-cloud/manage-identities-and-access).",
				Optional:            true,
			},
			"automq_byoc_secret_key": schema.StringAttribute{
				MarkdownDescription: "Set the Secret Access Key of client. AutoMQ Cloud (BYOC) requires Access Keys to manage access and authentication to different parts of the service. An Access Key consists of an access key id and a secret key. You can create and manage Access Keys by using the AutoMQ Cloud BYOC Console. Learn more about AutoMQ Cloud BYOC Console access [here](https://docs.automq.com/automq-cloud/manage-identities-and-access).",
				Optional:            true,
			},
			"automq_byoc_host": schema.StringAttribute{
				MarkdownDescription: "Set the AutoMQ BYOC environment endpoint. The endpoint like http://{hostname}:8080. You can get this endpoint when deploy environment complete.",
				Optional:            true,
			},
		},
	}
}

func (p *AutoMQProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring AutoMQ client")
	var data autoMQProviderModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Configuration values are now available.
	if data.BYOCHost.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("automq_byoc_host"),
			"Unknown AutoMQ API Host",
			"The provider cannot create the AutoMQ API client as there is an unknown configuration value for the AutoMQ API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTOMQ_HOST environment variable.",
		)
	}

	if data.BYOCAccessKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("automq_byoc_access_key_id"),
			"Unknown AutoMQ API BYOCAccessKey",
			"The provider cannot create the AutoMQ API client as there is an unknown configuration value for the AutoMQ API byoc_access_key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTOMQ_BYOC_ACCESS_KEY environment variable.",
		)
	}
	if data.BYOCSecretKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("automq_byoc_secret_key"),
			"Unknown AutoMQ API BYOCSecretKey",
			"The provider cannot create the AutoMQ API client as there is an unknown configuration value for the AutoMQ API byoc_secret_key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTOMQ_BYOC_SECRET_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	byco_host := os.Getenv("AUTOMQ_BYOC_HOST")
	byoc_access_key := os.Getenv("AUTOMQ_BYOC_ACCESS_KEY")
	byoc_secret_key := os.Getenv("AUTOMQ_BYOC_SECRET_KEY")

	if !data.BYOCHost.IsNull() {
		byco_host = data.BYOCHost.ValueString()
	}
	if !data.BYOCAccessKey.IsNull() {
		byoc_access_key = data.BYOCAccessKey.ValueString()
	}
	if !data.BYOCSecretKey.IsNull() {
		byoc_secret_key = data.BYOCSecretKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if byco_host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("automq_byoc_host"),
			"Missing AutoMQ API Host",
			"The provider cannot create the AutoMQ API client as there is a missing or empty value for the AutoMQ API host. "+
				"Set the host value in the configuration or use the AUTOMQ_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if byoc_access_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("automq_byoc_access_key_id"),
			"Missing AutoMQ API BYOCAccessKey",
			"The provider cannot create the AutoMQ API client as there is a missing or empty value for the AutoMQ API byoc_access_key. "+
				"Set the byoc_access_key value in the configuration or use the AUTOMQ_BYOC_ACCESS_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if byoc_secret_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("automq_byoc_secret_key"),
			"Missing AutoMQ API BYOCSecretKey",
			"The provider cannot create the AutoMQ API client as there is a missing or empty value for the AutoMQ API byoc_secret_key. "+
				"Set the byoc_secret_key value in the configuration or use the AUTOMQ_BYOC_SECRET_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "automq_env_byoc_host", byco_host)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "automq_env_token")

	tflog.Debug(ctx, "Creating AutoMQ client")

	credential := client.AuthCredentials{
		AccessKeyID:     data.BYOCAccessKey.ValueString(),
		SecretAccessKey: data.BYOCSecretKey.ValueString(),
	}

	client, err := client.NewClient(ctx, byco_host, credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AutoMQ API Client",
			"An unexpected error occurred when creating the AutoMQ API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"AutoMQ Client Error: "+err.Error(),
		)
		return
	}

	// Make the AutoMQ client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured AutoMQ client", map[string]any{"success": true})
}

func (p *AutoMQProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewKafkaInstanceResource,
		NewKafkaTopicResource,
		NewKafkaUserResource,
		NewKafkaAclResource,
		NewIntegrationResource,
	}
}

func (p *AutoMQProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewKafkaInstanceDataSource,
	}
}

func (p *AutoMQProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AutoMQProvider{
			version: version,
		}
	}
}
