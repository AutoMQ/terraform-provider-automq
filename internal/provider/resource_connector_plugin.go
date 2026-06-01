package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/framework"
	"terraform-provider-automq/internal/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	pluginDefaultCreateTimeout = 10 * time.Minute
	pluginDefaultDeleteTimeout = 10 * time.Minute
)

var (
	_ resource.Resource                = &ConnectorPluginResource{}
	_ resource.ResourceWithConfigure   = &ConnectorPluginResource{}
	_ resource.ResourceWithImportState = &ConnectorPluginResource{}
)

func NewConnectorPluginResource() resource.Resource {
	r := &ConnectorPluginResource{}
	r.SetDefaultCreateTimeout(pluginDefaultCreateTimeout)
	r.SetDefaultDeleteTimeout(pluginDefaultDeleteTimeout)
	return r
}

type ConnectorPluginResource struct {
	client *client.Client
	api    connectorPluginAPI
	framework.WithTimeouts
}

type connectorPluginAPI interface {
	CreateConnectPlugin(ctx context.Context, param client.ConnectPluginCreateParam) (*client.ConnectPluginVO, error)
	GetConnectPlugin(ctx context.Context, pluginId string) (*client.ConnectPluginVO, error)
	DeleteConnectPlugin(ctx context.Context, pluginId string) error
}

type defaultConnectorPluginAPI struct{ client *client.Client }

func (a defaultConnectorPluginAPI) CreateConnectPlugin(ctx context.Context, param client.ConnectPluginCreateParam) (*client.ConnectPluginVO, error) {
	return a.client.CreateConnectPlugin(ctx, param)
}
func (a defaultConnectorPluginAPI) GetConnectPlugin(ctx context.Context, id string) (*client.ConnectPluginVO, error) {
	return a.client.GetConnectPlugin(ctx, id)
}
func (a defaultConnectorPluginAPI) DeleteConnectPlugin(ctx context.Context, id string) error {
	return a.client.DeleteConnectPlugin(ctx, id)
}

func (r *ConnectorPluginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_plugin"
}

func (r *ConnectorPluginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
	r.api = defaultConnectorPluginAPI{client: c}
}

func (r *ConnectorPluginResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)\n\n" +
			"`automq_connector_plugin` registers a custom Kafka Connect plugin. Plugins are immutable after creation — to change a plugin, delete and recreate it.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Required:      true,
				Description:   "Target AutoMQ environment identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Plugin identifier assigned by the backend (e.g. `conn-plugin-xxxxx`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "Display name for the plugin.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"version": schema.StringAttribute{
				Required:      true,
				Description:   "Plugin version string (e.g. `1.0.0`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"storage_url": schema.StringAttribute{
				Required:      true,
				Description:   "URL where the plugin archive is stored. Supports `s3://`, `http://`, or `https://` schemes.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"types": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Plugin types: `SOURCE`, `SINK`, or both.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("SOURCE", "SINK")),
				},
				PlanModifiers: []planmodifier.List{},
			},
			"connector_class": schema.StringAttribute{
				Required:      true,
				Description:   "Fully-qualified Java class name of the connector, e.g. `io.confluent.connect.s3.S3SinkConnector`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Optional:      true,
				Description:   "Free-form text description of the plugin.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"documentation_link": schema.StringAttribute{
				Optional:      true,
				Description:   "URL to the plugin documentation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			// ---- computed / read-only ----
			"plugin_provider": schema.StringAttribute{
				Computed:      true,
				Description:   "Plugin provider: `AUTOMQ` (system built-in) or `CUSTOM` (user uploaded).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:      true,
				Description:   "Current plugin status: `ACTIVE`, `DISABLED`, `PENDING`, `DELETING`, or `DELETED`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "Timestamp when the plugin was created (RFC 3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "Timestamp of the last update to the plugin (RFC 3339).",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Delete: true}),
		},
	}
}

// ---------------------------------------------------------------------------
// CRUD (Create, Read, Delete — no Update since plugins are immutable)
// ---------------------------------------------------------------------------

func (r *ConnectorPluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ConnectorPluginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	request, diags := models.ExpandConnectorPluginCreate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.api.CreateConnectPlugin(ctx, *request)
	if err != nil {
		resp.Diagnostics.AddError("Create Connector Plugin Error", fmt.Sprintf("Unable to create connector plugin: %s", err))
		return
	}
	pluginID := derefString(created.Id)
	if pluginID == "" {
		resp.Diagnostics.AddError("Create Connector Plugin Error", "API returned empty plugin id")
		return
	}

	if err := waitForPluginActive(ctx, r.api, pluginID, r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connector Plugin Provisioning Error", err.Error())
		return
	}

	latest, err := r.api.GetConnectPlugin(ctx, pluginID)
	if err != nil {
		resp.Diagnostics.AddError("Read Connector Plugin Error", fmt.Sprintf("Unable to read plugin %q after create: %s", pluginID, err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnectorPlugin(latest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectorPluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ConnectorPluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	plugin, err := r.api.GetConnectPlugin(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Connector Plugin Error", fmt.Sprintf("Unable to read plugin %q: %s", state.ID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnectorPlugin(plugin, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConnectorPluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Plugins are immutable — all attributes have RequiresReplace, so Update should never be called.
	resp.Diagnostics.AddError("Update Not Supported", "Connector plugins are immutable. Delete and recreate to change.")
}

func (r *ConnectorPluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ConnectorPluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())
	pluginID := state.ID.ValueString()

	if err := r.api.DeleteConnectPlugin(ctx, pluginID); err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete Connector Plugin Error", fmt.Sprintf("Unable to delete plugin %q: %s", pluginID, err))
		return
	}
	if err := waitForPluginDeletion(ctx, r.api, pluginID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connector Plugin Delete Error", err.Error())
	}
}

func (r *ConnectorPluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Use <environment_id>@<plugin_id>.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// ---------------------------------------------------------------------------
// Wait helpers
// ---------------------------------------------------------------------------

func waitForPluginActive(ctx context.Context, api connectorPluginAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.PluginStatePending},
		Target:       []string{client.PluginStateActive},
		Refresh:      pluginStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        2 * time.Second,
		PollInterval: 5 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	return err
}

func waitForPluginDeletion(ctx context.Context, api connectorPluginAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.PluginStateActive, client.PluginStateDisabled, client.PluginStateDeleting},
		Target:       []string{client.PluginStateDeleted},
		Refresh:      pluginStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        2 * time.Second,
		PollInterval: 5 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	if ue, ok := err.(*retry.UnexpectedStateError); ok && ue.LastError == nil {
		return fmt.Errorf("plugin %q entered unexpected state %q while deleting", id, ue.State)
	}
	return err
}

func pluginStatusFunc(ctx context.Context, api connectorPluginAPI, id string) retry.StateRefreshFunc {
	var last string
	return func() (interface{}, string, error) {
		vo, err := api.GetConnectPlugin(ctx, id)
		if err != nil {
			if framework.IsNotFoundError(err) {
				return &client.ConnectPluginVO{}, client.PluginStateDeleted, nil
			}
			tflog.Warn(ctx, fmt.Sprintf("error refreshing plugin %q: %s", id, err))
			return nil, "", err
		}
		if vo == nil || vo.Status == nil {
			return nil, "", fmt.Errorf("plugin %q returned nil status", id)
		}
		cur := strings.ToUpper(*vo.Status)
		if cur != last {
			tflog.Debug(ctx, fmt.Sprintf("plugin %q status -> %s", id, cur))
			last = cur
		}
		return vo, cur, nil
	}
}
