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
	BYOCAccessKey types.String `tfsdk:"byoc_access_key"`
	BYOCSecretKey types.String `tfsdk:"byoc_secret_key"`
	BYOCHost      types.String `tfsdk:"byoc_host"`
	Token         types.String `tfsdk:"token"`
}

func (p *AutoMQProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "automq"
	resp.Version = p.version
}

func (p *AutoMQProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"byoc_access_key": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
			"byoc_secret_key": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
			"byoc_host": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
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
			path.Root("byoc_host"),
			"Unknown AutoMQ API Host",
			"The provider cannot create the AutoMQ API client as there is an unknown configuration value for the AutoMQ API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTOMQ_HOST environment variable.",
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown AutoMQ API Token",
			"The provider cannot create the AutoMQ API client as there is an unknown configuration value for the AutoMQ API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AUTOMQ_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	byco_host := os.Getenv("AUTOMQ_BYOC_HOST")
	token := os.Getenv("AUTOMQ_TOKEN")

	if !data.BYOCHost.IsNull() {
		byco_host = data.BYOCHost.ValueString()
	}
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if byco_host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("byoc_host"),
			"Missing AutoMQ API Host",
			"The provider cannot create the AutoMQ API client as there is a missing or empty value for the AutoMQ API host. "+
				"Set the host value in the configuration or use the AUTOMQ_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing AutoMQ API Token",
			"The provider cannot create the AutoMQ API client as there is a missing or empty value for the AutoMQ API token. "+
				"Set the token value in the configuration or use the AUTOMQ_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "automq_env_byoc_host", byco_host)
	ctx = tflog.SetField(ctx, "automq_env_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "automq_env_token")

	tflog.Debug(ctx, "Creating AutoMQ client")

	client, err := client.NewClient(ctx, &byco_host, &token)
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
	return []func() datasource.DataSource{}
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
