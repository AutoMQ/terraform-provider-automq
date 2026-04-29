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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	connectClusterDefaultCreateTimeout = 30 * time.Minute
	connectClusterDefaultUpdateTimeout = 30 * time.Minute
	connectClusterDefaultDeleteTimeout = 20 * time.Minute
)

var (
	workerResourceSpecs = []string{"TIER1", "TIER2", "TIER3", "TIER4"}
	capacityTypes       = []string{"provisioned", "autoscaling"}
	computeTypes        = []string{"k8s", "asg"}
)

var (
	_ resource.Resource                = &ConnectClusterResource{}
	_ resource.ResourceWithConfigure   = &ConnectClusterResource{}
	_ resource.ResourceWithImportState = &ConnectClusterResource{}
)

func NewConnectClusterResource() resource.Resource {
	r := &ConnectClusterResource{}
	r.SetDefaultCreateTimeout(connectClusterDefaultCreateTimeout)
	r.SetDefaultUpdateTimeout(connectClusterDefaultUpdateTimeout)
	r.SetDefaultDeleteTimeout(connectClusterDefaultDeleteTimeout)
	return r
}

type ConnectClusterResource struct {
	client *client.Client
	api    connectClusterAPI
	framework.WithTimeouts
}

type connectClusterAPI interface {
	CreateConnectCluster(ctx context.Context, param client.ConnectClusterCreateParam) (*client.ConnectClusterVO, error)
	GetConnectCluster(ctx context.Context, clusterId string) (*client.ConnectClusterVO, error)
	UpdateConnectCluster(ctx context.Context, clusterId string, param client.ConnectClusterUpdateParam) (*client.ConnectClusterVO, error)
	DeleteConnectCluster(ctx context.Context, clusterId string) error
}

type defaultConnectClusterAPI struct{ client *client.Client }

func (a defaultConnectClusterAPI) CreateConnectCluster(ctx context.Context, param client.ConnectClusterCreateParam) (*client.ConnectClusterVO, error) {
	return a.client.CreateConnectCluster(ctx, param)
}
func (a defaultConnectClusterAPI) GetConnectCluster(ctx context.Context, id string) (*client.ConnectClusterVO, error) {
	return a.client.GetConnectCluster(ctx, id)
}
func (a defaultConnectClusterAPI) UpdateConnectCluster(ctx context.Context, id string, param client.ConnectClusterUpdateParam) (*client.ConnectClusterVO, error) {
	return a.client.UpdateConnectCluster(ctx, id, param)
}
func (a defaultConnectClusterAPI) DeleteConnectCluster(ctx context.Context, id string) error {
	return a.client.DeleteConnectCluster(ctx, id)
}

func (r *ConnectClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connect_cluster"
}

func (r *ConnectClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
	r.api = defaultConnectClusterAPI{client: c}
}

func (r *ConnectClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`automq_connect_cluster` manages a Kafka Connect Worker cluster.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{Required: true, Description: "AutoMQ environment ID that owns the Connect Cluster, for example `env-xxxxx`. Changing it creates a new Connect Cluster.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"id":             schema.StringAttribute{Computed: true, Description: "Connect Cluster ID assigned by AutoMQ, for example `connect-cluster-xxxxx`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":           schema.StringAttribute{Required: true, Description: "Connect Cluster name shown in AutoMQ. It must be 3 to 64 characters.", Validators: []validator.String{stringvalidator.LengthBetween(3, 64)}},
			"description":    schema.StringAttribute{Optional: true, Description: "Optional human-readable description for the Connect Cluster."},
			"plugins": schema.ListNestedAttribute{
				Required:    true,
				Description: "Set of connector plugins installed into the worker cluster. Connectors on this cluster can only use connector classes provided by these plugins.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"name":    schema.StringAttribute{Required: true, Description: "Plugin name registered in AutoMQ, such as `s3-sink` or `snowflake`."},
					"version": schema.StringAttribute{Required: true, Description: "Plugin version registered in AutoMQ. The name and version pair must exist in the AutoMQ plugin repository."},
				}},
			},
			"kafka_cluster": schema.SingleNestedAttribute{
				Required:    true,
				Description: "AutoMQ Kafka instance used by the Connect workers for internal Connect topics and worker coordination.",
				Attributes: map[string]schema.Attribute{
					"kafka_instance_id": schema.StringAttribute{
						Required:      true,
						Description:   "AutoMQ Kafka instance ID, for example `kf-xxxxx`. Changing it creates a new Connect Cluster.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
				},
			},
			"capacity": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Worker capacity configuration for the Connect Cluster. Set `type` to `provisioned` for a fixed worker count or `autoscaling` for backend-managed scaling.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Required: true, Description: "Capacity mode. Supported values are `provisioned` and `autoscaling`.", Validators: []validator.String{stringvalidator.OneOf(capacityTypes...)}},
					"provisioned": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Fixed capacity settings. Use this block when `type` is `provisioned`.",
						Attributes: map[string]schema.Attribute{
							"worker_resource_spec": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf(workerResourceSpecs...)}, Description: "Worker resource tier. Supported values are `TIER1`, `TIER2`, `TIER3`, and `TIER4`."},
							"worker_count":         schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, Description: "Fixed number of Connect worker processes. Must be at least 1."},
						},
					},
					"autoscaling": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Autoscaling capacity settings. Use this block when `type` is `autoscaling`; backend support depends on the target AutoMQ control plane.",
						Attributes: map[string]schema.Attribute{
							"worker_resource_spec": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf(workerResourceSpecs...)}, Description: "Worker resource tier used by autoscaled workers. Supported values are `TIER1`, `TIER2`, `TIER3`, and `TIER4`."},
							"min_worker_count":     schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, Description: "Minimum number of Connect workers. Must be at least 1."},
							"max_worker_count":     schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, Description: "Maximum number of Connect workers. Must be greater than or equal to `min_worker_count`."},
							"scale_in_policy": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"cpu_utilization_percentage": schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 100)}, Description: "Average CPU utilization threshold, from 1 to 100, below which AutoMQ may scale workers in."},
								},
							},
							"scale_out_policy": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"cpu_utilization_percentage": schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 100)}, Description: "Average CPU utilization threshold, from 1 to 100, above which AutoMQ may scale workers out."},
								},
							},
						},
					},
				},
			},
			"compute": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Compute backend where Connect workers run.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Required: true, Description: "Compute type. Supported values are `k8s` and `asg`; `k8s` requires the `kubernetes` block. Changing it creates a new Connect Cluster.", Validators: []validator.String{stringvalidator.OneOf(computeTypes...)}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
					"kubernetes": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Kubernetes compute settings for worker pods. Required when `compute.type` is `k8s`.",
						Attributes: map[string]schema.Attribute{
							"cluster_id":      schema.StringAttribute{Required: true, Description: "AutoMQ-registered Kubernetes cluster ID where worker pods are deployed. Changing it creates a new Connect Cluster.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
							"namespace":       schema.StringAttribute{Required: true, Description: "Kubernetes namespace for worker pods and related resources. Changing it creates a new Connect Cluster.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
							"service_account": schema.StringAttribute{Required: true, Description: "Kubernetes ServiceAccount used by worker pods. In AWS environments this is typically bound to an IAM role by IRSA. Changing it creates a new Connect Cluster.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
							"scheduling_spec": schema.StringAttribute{Optional: true, Computed: true, Description: "Optional Kubernetes scheduling YAML, such as node selectors, tolerations, or affinity rules. If omitted, AutoMQ keeps the backend default.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
						},
					},
					"iam_role": schema.StringAttribute{Optional: true, Description: "Cloud IAM role used by Connect workers to access external services such as object storage. In AWS K8s deployments this is the IRSA role. Changing it creates a new Connect Cluster.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
				},
			},
			"worker_config":   schema.MapAttribute{ElementType: types.StringType, Optional: true, Computed: true, Description: "Worker-level Kafka Connect configuration overrides, such as converters or offset flush settings. Do not put plugin-specific connector configuration here; use `automq_connector.connector_config` instead.", PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()}},
			"metric_exporter": metricExporterSchema(),
			"tags":            schema.MapAttribute{ElementType: types.StringType, Optional: true, Computed: true, Description: "User-defined tags for organizing Connect Clusters. AutoMQ may also return system tags.", PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()}},
			"version":         schema.StringAttribute{Optional: true, Computed: true, Description: "AutoMQ Connect worker version. If omitted, AutoMQ selects the backend default version.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"state":           schema.StringAttribute{Computed: true, Description: "Connect Cluster lifecycle state reported by AutoMQ, such as `RUNNING`, `CHANGING`, `FAILED`, or `DELETING`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"kafka_connect_version": schema.StringAttribute{
				Computed:      true,
				Description:   "Kafka Connect framework version used by the worker runtime.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{Computed: true, CustomType: timetypes.RFC3339Type{}, Description: "Creation timestamp."},
			"updated_at": schema.StringAttribute{Computed: true, CustomType: timetypes.RFC3339Type{}, Description: "Last update timestamp."},
			"timeouts":   timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func metricExporterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Metrics exporter configuration for Connect workers.",
		Attributes: map[string]schema.Attribute{
			"remote_write": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Prometheus Remote Write exporter configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled":              schema.BoolAttribute{Optional: true, Description: "Whether remote write is enabled."},
					"endpoint":             schema.StringAttribute{Optional: true, Description: "Prometheus Remote Write endpoint URL."},
					"auth_type":            schema.StringAttribute{Optional: true, Description: "Authentication type for the remote write endpoint, such as basic, bearer token, or AWS SigV4 depending on backend support."},
					"username":             schema.StringAttribute{Optional: true, Description: "Basic auth username."},
					"password":             schema.StringAttribute{Optional: true, Sensitive: true, Description: "Basic auth password."},
					"token":                schema.StringAttribute{Optional: true, Sensitive: true, Description: "Bearer token."},
					"region":               schema.StringAttribute{Optional: true, Description: "AWS SigV4 region."},
					"prometheus_arn":       schema.StringAttribute{Optional: true, Description: "AWS Managed Service for Prometheus workspace ARN used with SigV4 authentication."},
					"insecure_skip_verify": schema.BoolAttribute{Optional: true, Description: "Whether to skip TLS certificate verification when sending metrics. Use only for controlled test environments."},
					"headers":              schema.MapAttribute{Optional: true, ElementType: types.StringType, Description: "Additional HTTP headers sent with remote write requests."},
					"labels":               schema.MapAttribute{Optional: true, ElementType: types.StringType, Description: "Additional labels attached to exported worker metrics."},
				},
			},
		},
	}
}

func (r *ConnectClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ConnectClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, plan.EnvironmentID.ValueString())

	request, diags := models.ExpandConnectClusterCreate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.api.CreateConnectCluster(ctx, *request)
	if err != nil {
		resp.Diagnostics.AddError("Create Connect Cluster Error", fmt.Sprintf("Unable to create connect cluster: %s", err))
		return
	}
	clusterID := derefString(created.Id)
	if clusterID == "" {
		resp.Diagnostics.AddError("Create Connect Cluster Error", "API returned empty connect cluster id")
		return
	}
	if err := waitForConnectClusterReady(ctx, r.api, clusterID, r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connect Cluster Provisioning Error", err.Error())
		return
	}
	latest, err := r.api.GetConnectCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError("Read Connect Cluster Error", fmt.Sprintf("Unable to read connect cluster %q after create: %s", clusterID, err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnectCluster(latest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ConnectClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())

	cluster, err := r.api.GetConnectCluster(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Connect Cluster Error", fmt.Sprintf("Unable to read connect cluster %q: %s", state.ID.ValueString(), err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnectCluster(cluster, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConnectClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.ConnectClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())
	clusterID := state.ID.ValueString()

	request, diags := models.ExpandConnectClusterUpdate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.api.UpdateConnectCluster(ctx, clusterID, *request); err != nil {
		resp.Diagnostics.AddError("Update Connect Cluster Error", fmt.Sprintf("Unable to update connect cluster %q: %s", clusterID, err))
		return
	}
	if err := waitForConnectClusterReady(ctx, r.api, clusterID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connect Cluster Update Error", err.Error())
		return
	}
	latest, err := r.api.GetConnectCluster(ctx, clusterID)
	if err != nil {
		resp.Diagnostics.AddError("Read Connect Cluster Error", fmt.Sprintf("Unable to refresh connect cluster %q: %s", clusterID, err))
		return
	}
	resp.Diagnostics.Append(models.FlattenConnectCluster(latest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ConnectClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx = context.WithValue(ctx, client.EnvIdKey, state.EnvironmentID.ValueString())
	clusterID := state.ID.ValueString()
	if err := r.api.DeleteConnectCluster(ctx, clusterID); err != nil {
		if framework.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete Connect Cluster Error", fmt.Sprintf("Unable to delete connect cluster %q: %s", clusterID, err))
		return
	}
	if err := waitForConnectClusterDeletion(ctx, r.api, clusterID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError("Connect Cluster Delete Error", err.Error())
	}
}

func (r *ConnectClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Use <environment_id>@<connect_cluster_id>.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func waitForConnectClusterReady(ctx context.Context, api connectClusterAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.ConnectClusterStateCreating, client.ConnectClusterStateChanging, client.ConnectClusterStateUnknown},
		Target:       []string{client.ConnectClusterStateRunning},
		Refresh:      connectClusterStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: 15 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	return err
}

func waitForConnectClusterDeletion(ctx context.Context, api connectClusterAPI, id string, timeout time.Duration) error {
	conf := &retry.StateChangeConf{
		Pending:      []string{client.ConnectClusterStateDeleting, client.ConnectClusterStateChanging, client.ConnectClusterStateUnknown},
		Target:       []string{client.ConnectClusterStateDeleted},
		Refresh:      connectClusterStatusFunc(ctx, api, id),
		Timeout:      timeout,
		Delay:        5 * time.Second,
		PollInterval: 10 * time.Second,
	}
	_, err := conf.WaitForStateContext(ctx)
	return err
}

func connectClusterStatusFunc(ctx context.Context, api connectClusterAPI, id string) retry.StateRefreshFunc {
	var last string
	return func() (interface{}, string, error) {
		vo, err := api.GetConnectCluster(ctx, id)
		if err != nil {
			if framework.IsNotFoundError(err) {
				return &client.ConnectClusterVO{}, client.ConnectClusterStateDeleted, nil
			}
			tflog.Warn(ctx, fmt.Sprintf("error refreshing connect cluster %q: %s", id, err))
			return nil, client.ConnectClusterStateUnknown, err
		}
		if vo == nil || vo.State == nil {
			return nil, client.ConnectClusterStateUnknown, fmt.Errorf("connect cluster %q returned nil state", id)
		}
		cur := strings.ToUpper(*vo.State)
		if cur != last {
			tflog.Debug(ctx, fmt.Sprintf("connect cluster %q state -> %s", id, cur))
			last = cur
		}
		if cur == client.ConnectClusterStateFailed {
			return vo, cur, fmt.Errorf("connect cluster %q entered FAILED state", id)
		}
		return vo, cur, nil
	}
}
