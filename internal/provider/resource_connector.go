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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	connectorDefaultCreateTimeout = 30 * time.Minute
	connectorDefaultUpdateTimeout = 30 * time.Minute
	connectorDefaultDeleteTimeout = 20 * time.Minute
)

var (
	workerResourceSpecs = []string{"TIER1", "TIER2", "TIER3", "TIER4"}
	pluginTypes         = []string{"SOURCE", "SINK"}
	securityProtocols   = []string{"PLAINTEXT", "SSL", "SASL_PLAINTEXT", "SASL_SSL"}
)

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
		MarkdownDescription: "![Preview](https://img.shields.io/badge/Lifecycle_Stage-Preview-blue?style=flat&logoColor=8A3BE2&labelColor=rgba)\n\n" +
			"`automq_connector` provisions and manages Kafka Connect clusters that run alongside your AutoMQ instances.",
		Attributes: map[string]schema.Attribute{
			// ---- identity / lifecycle ----
			"environment_id": schema.StringAttribute{
				Required:      true,
				Description:   "Target AutoMQ environment identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Connector cluster identifier, assigned by the backend (e.g. `conn-xxxxx`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},

			// ---- user-configurable ----
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the connector cluster. Must be between 3 and 64 characters.",
				Validators:  []validator.String{stringvalidator.LengthBetween(3, 64)},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Free-form text description of the connector cluster.",
			},
			"plugin_id": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of a registered Connect plugin (e.g. `conn-plugin-xxxxx`). The plugin determines which connector classes are available.",
			},
			"plugin_type": schema.StringAttribute{
				Optional:      true,
				Description:   "Plugin type hint (`SOURCE` or `SINK`).",
				Validators:    []validator.String{stringvalidator.OneOf(pluginTypes...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"connector_class": schema.StringAttribute{
				Required:      true,
				Description:   "Fully-qualified Java class name of the connector, e.g. `io.confluent.connect.s3.S3SinkConnector`. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"iam_role": schema.StringAttribute{
				Optional:      true,
				Description:   "AWS IAM Role ARN for the connector pods (IRSA). Grants the connector access to AWS resources such as S3 buckets. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kubernetes_cluster_id": schema.StringAttribute{
				Required:      true,
				Description:   "Target Kubernetes cluster ID where connector pods will be deployed. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kubernetes_namespace": schema.StringAttribute{
				Required:      true,
				Description:   "Kubernetes namespace in which connector pods run. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"kubernetes_service_account": schema.StringAttribute{
				Required:      true,
				Description:   "Kubernetes ServiceAccount used by connector pods. Typically associated with an IAM role via IRSA for cloud resource access. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"task_count": schema.Int64Attribute{
				Required:    true,
				Description: "Number of connector tasks to run in parallel. Each task handles a portion of the data. Must be at least 1.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"version": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "AutoMQ version for the connector cluster. If omitted, the backend selects the latest available version. Can be updated later to trigger a rolling upgrade.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"scheduling_spec": schema.StringAttribute{
				Optional:    true,
				Description: "Kubernetes scheduling spec as a YAML snippet. Supports nodeSelector, tolerations, and affinity rules for controlling pod placement.",
			},
			"worker_config": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				Description:   "Worker-level Kafka Connect configuration overrides (key-value pairs). Controls internal topics, serialization defaults, and other worker bootstrap settings such as `offset.storage.topic` or `key.converter`.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"connector_config": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				Description:   "Connector-level configuration pushed via the Connect REST API. Contains plugin-specific settings such as `topics`, `s3.bucket.name`, `flush.size`, etc.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"labels": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				Description:   "Custom key-value labels for organizing and filtering connector clusters.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},

			// ---- nested blocks ----
			"capacity": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Worker capacity configuration that determines the compute resources allocated to the connector cluster.",
				Attributes: map[string]schema.Attribute{
					"worker_count": schema.Int64Attribute{
						Required:    true,
						Description: "Number of Kafka Connect worker pods. Each worker is an independent Connect process. Must be at least 1.",
						Validators:  []validator.Int64{int64validator.AtLeast(1)},
					},
					"worker_resource_spec": schema.StringAttribute{
						Required:    true,
						Description: "Worker pod resource tier. `TIER1` (0.5 CPU / 1 GiB), `TIER2` (1 CPU / 2 GiB), `TIER3` (2 CPU / 4 GiB), `TIER4` (4 CPU / 8 GiB).",
						Validators:  []validator.String{stringvalidator.OneOf(workerResourceSpecs...)},
					},
				},
			},
			"kafka_cluster": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Kafka cluster binding. Specifies which AutoMQ Kafka instance the connector connects to and how it authenticates.",
				Attributes: map[string]schema.Attribute{
					"kafka_instance_id": schema.StringAttribute{
						Required:      true,
						Description:   "AutoMQ Kafka instance ID to bind to (e.g. `kf-xxxxx`). Changing this forces a new resource.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"security_protocol": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Security protocol configuration for the Kafka connection. Required fields depend on the chosen protocol: PLAINTEXT needs nothing extra; SASL_PLAINTEXT/SASL_SSL need username+password; SSL (mTLS) needs client_cert+private_key.",
						Attributes: map[string]schema.Attribute{
							"security_protocol": schema.StringAttribute{
								Required:    true,
								Description: "Protocol type: `PLAINTEXT` (no auth), `SASL_PLAINTEXT` (SASL over plaintext), `SASL_SSL` (SASL over TLS), or `SSL` (mTLS mutual authentication).",
								Validators:  []validator.String{stringvalidator.OneOf(securityProtocols...)},
							},
							"username": schema.StringAttribute{
								Optional:    true,
								Description: "SASL username. Required when security_protocol is `SASL_PLAINTEXT` or `SASL_SSL`.",
							},
							"password": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "SASL password. Required when security_protocol is `SASL_PLAINTEXT` or `SASL_SSL`.",
							},
							"sasl_mechanism": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								Description:   "SASL mechanism, e.g. `PLAIN`, `SCRAM-SHA-256`, `SCRAM-SHA-512`. Defaults to `SCRAM-SHA-512` on the server side if omitted.",
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"truststore_certs": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								Description:   "Custom CA certificate (PEM format) for TLS verification. Applicable to `SSL` and `SASL_SSL`. If omitted, the CA certificate from the Kafka instance is used automatically.",
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"key_password": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "Passphrase for the encrypted private key. Only required if `private_key` is password-protected.",
							},
							"client_cert": schema.StringAttribute{
								Optional:    true,
								Description: "Client certificate in PEM format for mTLS authentication. Required when security_protocol is `SSL`.",
							},
							"private_key": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "Client private key in PEM format for mTLS authentication. Required when security_protocol is `SSL`.",
							},
						},
					},
				},
			},
			"metric_exporter": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Metrics exporter configuration for the connector cluster.",
				Attributes: map[string]schema.Attribute{
					"remote_write": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Prometheus Remote Write configuration. Pushes connector metrics to an external Prometheus-compatible endpoint.",
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether Prometheus remote write is enabled.",
							},
							"endpoint": schema.StringAttribute{
								Optional:    true,
								Description: "Remote write endpoint URL, e.g. `https://prometheus.example.com/api/v1/write`.",
							},
							"auth_type": schema.StringAttribute{
								Optional:    true,
								Description: "Authentication type for the remote write endpoint: `none`, `basic`, `bearer`, or `sigv4`.",
							},
							"username": schema.StringAttribute{
								Optional:    true,
								Description: "Username for basic authentication (`auth_type = basic`).",
							},
							"password": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "Password for basic authentication (`auth_type = basic`).",
							},
							"token": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "Bearer token for token-based authentication (`auth_type = bearer`).",
							},
							"region": schema.StringAttribute{
								Optional:    true,
								Description: "AWS region for SigV4 authentication (`auth_type = sigv4`).",
							},
							"prometheus_arn": schema.StringAttribute{
								Optional:    true,
								Description: "Amazon Managed Prometheus workspace ARN for SigV4 authentication (`auth_type = sigv4`).",
							},
							"insecure_skip_verify": schema.BoolAttribute{
								Optional:    true,
								Description: "Skip TLS certificate verification for the remote write endpoint. Not recommended for production.",
							},
							"headers": schema.MapAttribute{
								Optional:    true,
								ElementType: types.StringType,
								Description: "Custom HTTP headers to include in remote write requests.",
							},
							"labels": schema.MapAttribute{
								Optional:    true,
								ElementType: types.StringType,
								Description: "Custom labels appended to all metrics sent via remote write.",
							},
						},
						PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
					},
				},
			},

			// ---- computed / read-only ----
			"state": schema.StringAttribute{
				Computed:      true,
				Description:   "Current lifecycle state of the connector cluster: `CREATING`, `RUNNING`, `PAUSED`, `CHANGING`, `FAILED`, `DELETING`, or `DELETED`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"kafka_connect_version": schema.StringAttribute{
				Computed:      true,
				Description:   "Kafka Connect framework version determined by the backend based on the AutoMQ version (e.g. `3.7.0`). Read-only.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "Timestamp when the connector cluster was created (RFC 3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "Timestamp of the last update to the connector cluster (RFC 3339).",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Wait helpers
// ---------------------------------------------------------------------------

func waitForConnectorReady(ctx context.Context, api connectorAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.ConnectorStateCreating, client.ConnectorStateChanging, client.ConnectorStateUnknown},
		Target:       []string{client.ConnectorStateRunning, client.ConnectorStatePaused},
		Refresh:      connectorStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: 15 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	return err
}

func waitForConnectorDeletion(ctx context.Context, api connectorAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.ConnectorStateDeleting, client.ConnectorStateChanging},
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
