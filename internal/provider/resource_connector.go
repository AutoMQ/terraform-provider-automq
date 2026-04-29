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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	connectorDefaultCreateTimeout = 10 * time.Minute
	connectorDefaultUpdateTimeout = 10 * time.Minute
	connectorDefaultDeleteTimeout = 10 * time.Minute
)

var securityProtocols = []string{"PLAINTEXT", "SSL", "SASL_PLAINTEXT", "SASL_SSL"}

var (
	_ resource.Resource                = &ConnectorResource{}
	_ resource.ResourceWithConfigure   = &ConnectorResource{}
	_ resource.ResourceWithImportState = &ConnectorResource{}
)

func NewConnectorResource() resource.Resource {
	r := &ConnectorResource{}
	r.SetDefaultCreateTimeout(connectorDefaultCreateTimeout)
	r.SetDefaultUpdateTimeout(connectorDefaultUpdateTimeout)
	r.SetDefaultDeleteTimeout(connectorDefaultDeleteTimeout)
	return r
}

type ConnectorResource struct {
	client *client.Client
	api    connectorAPI
	framework.WithTimeouts
}

type connectorAPI interface {
	CreateConnector(ctx context.Context, param client.ConnectorCreateParam) (*client.ConnectorVO, error)
	GetConnector(ctx context.Context, connectorId string) (*client.ConnectorVO, error)
	UpdateConnector(ctx context.Context, connectorId string, param client.ConnectorUpdateParam) (*client.ConnectorVO, error)
	DeleteConnector(ctx context.Context, connectorId string) error
}

type defaultConnectorAPI struct{ client *client.Client }

func (a defaultConnectorAPI) CreateConnector(ctx context.Context, param client.ConnectorCreateParam) (*client.ConnectorVO, error) {
	return a.client.CreateConnector(ctx, param)
}
func (a defaultConnectorAPI) GetConnector(ctx context.Context, id string) (*client.ConnectorVO, error) {
	return a.client.GetConnector(ctx, id)
}
func (a defaultConnectorAPI) UpdateConnector(ctx context.Context, id string, param client.ConnectorUpdateParam) (*client.ConnectorVO, error) {
	return a.client.UpdateConnector(ctx, id, param)
}
func (a defaultConnectorAPI) DeleteConnector(ctx context.Context, id string) error {
	return a.client.DeleteConnector(ctx, id)
}

func (r *ConnectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (r *ConnectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
	r.api = defaultConnectorAPI{client: c}
}

func (r *ConnectorResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`automq_connector` manages a Kafka Connector instance on an existing AutoMQ Connect cluster.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Required:      true,
				Description:   "AutoMQ environment ID that owns the Connect resources, for example `env-xxxxx`. Changing it creates a new connector.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Connector ID assigned by AutoMQ, for example `conn-xxxxx`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"connect_cluster_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the `automq_connect_cluster` that hosts this connector. The target cluster must already have a plugin that provides `connector_class`. Changing this value creates a new connector.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Globally unique connector name shown in AutoMQ. It must be 3 to 64 characters.",
				Validators:  []validator.String{stringvalidator.LengthBetween(3, 64)},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional human-readable description for the connector.",
			},
			"connector_class": schema.StringAttribute{
				Required:      true,
				Description:   "Fully-qualified Kafka Connect connector class, such as `io.confluent.connect.s3.S3SinkConnector`. AutoMQ resolves the plugin from the target Connect Cluster by this class. Changing it creates a new connector.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"task_count": schema.Int64Attribute{
				Required:    true,
				Description: "Maximum number of Kafka Connect tasks for this connector. This maps to Connect `tasks.max` and must be at least 1.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"connector_config": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				Description:   "Plugin-specific non-sensitive connector configuration as Kafka Connect key-value properties, for example `topics`, `s3.bucket.name`, or `flush.size`. AutoMQ injects `connector.class` and `tasks.max`; do not set them here.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"connector_config_sensitive": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Sensitive:     true,
				Description:   "Plugin-specific sensitive connector configuration, such as passwords, tokens, and private keys. Values are marked sensitive in Terraform and retained in state when the API masks them on read.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"kafka_cluster": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Kafka client authentication used by the connector plugin's producer or consumer clients. Worker-level Kafka authentication is managed by AutoMQ and is not configured here.",
				Attributes: map[string]schema.Attribute{
					"security_protocol": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Security protocol settings injected into the connector's producer and consumer override configuration.",
						Attributes: map[string]schema.Attribute{
							"protocol": schema.StringAttribute{
								Required:    true,
								Description: "Kafka security protocol. Supported values are `PLAINTEXT`, `SSL`, `SASL_PLAINTEXT`, and `SASL_SSL`.",
								Validators:  []validator.String{stringvalidator.OneOf(securityProtocols...)},
							},
							"username": schema.StringAttribute{Optional: true, Description: "SASL username. Required by the backend connector runtime when `protocol` uses SASL."},
							"password": schema.StringAttribute{Optional: true, Sensitive: true, Description: "SASL password. Required by the backend connector runtime when `protocol` uses SASL."},
							"sasl_mechanism": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								Description:   "SASL mechanism, such as `SCRAM-SHA-512`. If omitted for a SASL protocol, AutoMQ defaults to `SCRAM-SHA-512`.",
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"truststore_certs": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								Description:   "Custom CA certificates in PEM format for SSL trust. If omitted, connector clients use the runtime default trust configuration.",
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"client_cert": schema.StringAttribute{Optional: true, Description: "Client certificate chain in PEM format for mTLS."},
							"private_key": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Client private key in PEM format for mTLS."},
						},
					},
				},
			},
			"initial_offsets": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Initial Kafka Connect offsets to apply when the connector is created. This is create-only; changing it creates a new connector.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"partition": schema.MapAttribute{Required: true, ElementType: types.StringType, Description: "Connector-specific source partition map used by Kafka Connect offsets."},
					"offset":    schema.MapAttribute{Required: true, ElementType: types.StringType, Description: "Connector-specific source offset map used by Kafka Connect offsets."},
				}},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"state": schema.StringAttribute{
				Computed:      true,
				Description:   "Connector lifecycle state reported by AutoMQ, such as `RUNNING`, `PAUSED`, `CHANGING`, `FAILED`, or `DELETING`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"connector_type": schema.StringAttribute{
				Computed:      true,
				Description:   "Connector type inferred by AutoMQ from the resolved plugin, such as `SOURCE` or `SINK`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"plugin_id": schema.StringAttribute{
				Computed:      true,
				Description:   "Resolved plugin ID. This is computed from `connect_cluster_id` and `connector_class`; it is not an input field.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "Creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "Last update timestamp.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *ConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	request, diags := models.ExpandConnectorCreate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.api.CreateConnector(ctx, *request)
	if err != nil {
		resp.Diagnostics.AddError("Create Connector Error", fmt.Sprintf("Unable to create connector: %s", err))
		return
	}
	connectorID := derefString(created.Id)
	if connectorID == "" {
		resp.Diagnostics.AddError("Create Connector Error", "API returned empty connector id")
		return
	}
	if err := waitForConnectorReady(ctx, r.api, connectorID, r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connector Provisioning Error", err.Error())
		return
	}
	latest, err := r.api.GetConnector(ctx, connectorID)
	if err != nil {
		resp.Diagnostics.AddError("Read Connector Error", fmt.Sprintf("Unable to read connector %q after create: %s", connectorID, err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnector(latest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	connector, err := r.api.GetConnector(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Connector Error", fmt.Sprintf("Unable to read connector %q: %s", state.ID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnector(connector, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.ConnectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())
	connectorID := state.ID.ValueString()

	request, diags := models.ExpandConnectorUpdate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.api.UpdateConnector(ctx, connectorID, *request); err != nil {
		resp.Diagnostics.AddError("Update Connector Error", fmt.Sprintf("Unable to update connector %q: %s", connectorID, err))
		return
	}
	if err := waitForConnectorReady(ctx, r.api, connectorID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connector Update Error", err.Error())
		return
	}
	latest, err := r.api.GetConnector(ctx, connectorID)
	if err != nil {
		resp.Diagnostics.AddError("Read Connector Error", fmt.Sprintf("Unable to refresh connector %q: %s", connectorID, err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnector(latest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ConnectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())
	connectorID := state.ID.ValueString()

	if err := r.api.DeleteConnector(ctx, connectorID); err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete Connector Error", fmt.Sprintf("Unable to delete connector %q: %s", connectorID, err))
		return
	}
	if err := waitForConnectorDeletion(ctx, r.api, connectorID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connector Delete Error", err.Error())
	}
}

func (r *ConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Use <environment_id>@<connector_id>.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func waitForConnectorReady(ctx context.Context, api connectorAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.ConnectorStateCreating, client.ConnectorStateChanging, client.ConnectorStateUnknown},
		Target:       []string{client.ConnectorStateRunning, client.ConnectorStatePaused},
		Refresh:      connectorStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        5 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	return err
}

func waitForConnectorDeletion(ctx context.Context, api connectorAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.ConnectorStateDeleting, client.ConnectorStateChanging, client.ConnectorStateUnknown},
		Target:       []string{client.ConnectorStateDeleted},
		Refresh:      connectorStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        5 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	if ue, ok := err.(*retry.UnexpectedStateError); ok && ue.LastError == nil {
		return fmt.Errorf("connector %q entered unexpected state %q while deleting", id, ue.State)
	}
	return err
}

func connectorStatusFunc(ctx context.Context, api connectorAPI, id string) retry.StateRefreshFunc {
	var last string
	return func() (interface{}, string, error) {
		vo, err := api.GetConnector(ctx, id)
		if err != nil {
			if framework.IsNotFoundError(err) {
				return &client.ConnectorVO{}, client.ConnectorStateDeleted, nil
			}
			tflog.Warn(ctx, fmt.Sprintf("error refreshing connector %q: %s", id, err))
			return nil, client.ConnectorStateUnknown, err
		}
		if vo == nil || vo.State == nil {
			return nil, client.ConnectorStateUnknown, fmt.Errorf("connector %q returned nil state", id)
		}
		cur := strings.ToUpper(*vo.State)
		if cur != last {
			tflog.Debug(ctx, fmt.Sprintf("connector %q state -> %s", id, cur))
			last = cur
		}
		if cur == client.ConnectorStateFailed {
			return vo, cur, fmt.Errorf("connector %q entered FAILED state", id)
		}
		return vo, cur, nil
	}
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
