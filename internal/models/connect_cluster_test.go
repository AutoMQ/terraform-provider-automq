package models

import (
	"context"
	"terraform-provider-automq/client"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandConnectClusterCreate(t *testing.T) {
	ctx := context.Background()
	workerConfig, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"offset.flush.interval.ms": "10000"})
	assert.False(t, diags.HasError())
	tags, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "data"})
	assert.False(t, diags.HasError())

	plan := ConnectClusterResourceModel{
		Name:        types.StringValue("cluster-a"),
		Description: types.StringValue("desc"),
		Plugins: []ConnectClusterPluginModel{
			{Name: types.StringValue("s3-sink"), Version: types.StringValue("11.1.0")},
		},
		KafkaCluster: &ConnectClusterKafkaModel{KafkaInstanceID: types.StringValue("kf-1")},
		Capacity: &ConnectClusterCapacityModel{
			Type: types.StringValue("provisioned"),
			Provisioned: &ConnectClusterProvisionedCapacityModel{
				WorkerResourceSpec: types.StringValue("TIER2"),
				WorkerCount:        types.Int64Value(2),
			},
		},
		Compute: &ConnectClusterComputeModel{
			Type:    types.StringValue("k8s"),
			IamRole: types.StringValue("arn:aws:iam::123:role/connect"),
			Kubernetes: &ConnectClusterKubernetesModel{
				ClusterID:      types.StringValue("eks-1"),
				Namespace:      types.StringValue("connect"),
				ServiceAccount: types.StringValue("connect-sa"),
			},
		},
		WorkerConfig: workerConfig,
		Tags:         tags,
		Version:      types.StringValue("5.3.8"),
	}

	req, diags := ExpandConnectClusterCreate(plan)
	assert.False(t, diags.HasError())
	assert.Equal(t, "cluster-a", req.Name)
	assert.Equal(t, "desc", *req.Description)
	assert.Equal(t, "s3-sink", req.Plugins[0].Name)
	assert.Equal(t, "11.1.0", req.Plugins[0].Version)
	assert.Equal(t, "kf-1", req.KafkaCluster.KafkaInstanceId)
	assert.Equal(t, "provisioned", req.Capacity.Type)
	assert.Equal(t, int32(2), req.Capacity.Provisioned.WorkerCount)
	assert.Equal(t, "TIER2", req.Capacity.Provisioned.WorkerResourceSpec)
	assert.Equal(t, "k8s", req.Compute.Type)
	assert.Equal(t, "eks-1", req.Compute.Kubernetes.ClusterId)
	assert.Equal(t, "10000", req.WorkerConfig.Properties["offset.flush.interval.ms"])
	assert.Equal(t, "data", req.Tags["team"])
	assert.Equal(t, "5.3.8", *req.Version)
}

func TestFlattenConnectCluster(t *testing.T) {
	now := time.Now()
	workerCount := int32(2)
	vo := &client.ConnectClusterVO{
		Id:                       connStrPtr("connect-cluster-1"),
		Name:                     connStrPtr("cluster-a"),
		Description:              connStrPtr("desc"),
		State:                    connStrPtr("RUNNING"),
		Plugins:                  []client.ClusterPluginVO{{Name: connStrPtr("s3-sink"), Version: connStrPtr("11.1.0")}},
		KafkaInstanceId:          connStrPtr("kf-1"),
		KubernetesClusterId:      connStrPtr("eks-1"),
		KubernetesNamespace:      connStrPtr("connect"),
		KubernetesServiceAccount: connStrPtr("connect-sa"),
		WorkerCount:              &workerCount,
		WorkerResourceSpec:       connStrPtr("TIER2"),
		CapacityType:             connStrPtr("PROVISIONED"),
		WorkerConfig:             map[string]interface{}{"offset.flush.interval.ms": "10000"},
		Tags:                     map[string]string{"team": "data"},
		Version:                  connStrPtr("5.3.8"),
		KafkaConnectVersion:      connStrPtr("3.9.0"),
		CreateTime:               &now,
		UpdateTime:               &now,
	}
	state := &ConnectClusterResourceModel{}

	diags := FlattenConnectCluster(vo, state)
	assert.False(t, diags.HasError())
	assert.Equal(t, "connect-cluster-1", state.ID.ValueString())
	assert.Equal(t, "cluster-a", state.Name.ValueString())
	assert.Equal(t, "RUNNING", state.State.ValueString())
	assert.Equal(t, "s3-sink", state.Plugins[0].Name.ValueString())
	assert.Equal(t, "kf-1", state.KafkaCluster.KafkaInstanceID.ValueString())
	assert.Equal(t, "eks-1", state.Compute.Kubernetes.ClusterID.ValueString())
	assert.Equal(t, "provisioned", state.Capacity.Type.ValueString())
	assert.Equal(t, int64(2), state.Capacity.Provisioned.WorkerCount.ValueInt64())
	assert.Equal(t, "3.9.0", state.KafkaConnectVersion.ValueString())
}

func TestFlattenConnectCluster_RetainsEmptyWorkerConfig(t *testing.T) {
	ctx := context.Background()
	emptyWorkerConfig, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{})
	assert.False(t, diags.HasError())
	state := &ConnectClusterResourceModel{WorkerConfig: emptyWorkerConfig}

	diags = FlattenConnectCluster(&client.ConnectClusterVO{Id: connStrPtr("connect-cluster-1")}, state)
	assert.False(t, diags.HasError())
	assert.False(t, state.WorkerConfig.IsNull())
	assert.Empty(t, state.WorkerConfig.Elements())
}

func TestConnectClusterSchedulingSpec_NormalizesEquivalentYAML(t *testing.T) {
	input := "nodeSelector:\n  dedicated: automq-connect\ntolerations:\n- key: dedicated\n  operator: Equal\n  value: automq-connect\n  effect: NoSchedule\n"
	api := "nodeSelector: {dedicated: automq-connect}\ntolerations:\n  - {key: dedicated, operator: Equal, value: automq-connect, effect: NoSchedule}\n"
	plan := ConnectClusterResourceModel{
		Name: types.StringValue("cluster-a"),
		Plugins: []ConnectClusterPluginModel{
			{Name: types.StringValue("s3-sink"), Version: types.StringValue("11.1.0")},
		},
		KafkaCluster: &ConnectClusterKafkaModel{KafkaInstanceID: types.StringValue("kf-1")},
		Capacity: &ConnectClusterCapacityModel{
			Type: types.StringValue("provisioned"),
			Provisioned: &ConnectClusterProvisionedCapacityModel{
				WorkerResourceSpec: types.StringValue("TIER1"),
				WorkerCount:        types.Int64Value(1),
			},
		},
		Compute: &ConnectClusterComputeModel{
			Type: types.StringValue("k8s"),
			Kubernetes: &ConnectClusterKubernetesModel{
				ClusterID:      types.StringValue("eks-1"),
				Namespace:      types.StringValue("connect"),
				ServiceAccount: types.StringValue("connect-sa"),
				SchedulingSpec: types.StringValue(input),
			},
		},
	}

	req, diags := ExpandConnectClusterCreate(plan)
	assert.False(t, diags.HasError())
	assert.NotNil(t, req.Compute.Kubernetes.SchedulingSpec)

	state := &ConnectClusterResourceModel{Compute: plan.Compute}
	diags = FlattenConnectCluster(&client.ConnectClusterVO{Id: connStrPtr("connect-cluster-1"), SchedulingSpec: &api}, state)
	assert.False(t, diags.HasError())
	assert.Equal(t, *req.Compute.Kubernetes.SchedulingSpec, state.Compute.Kubernetes.SchedulingSpec.ValueString())
}

func TestConnectClusterSchedulingSpec_EmptyYamlIsOmitted(t *testing.T) {
	for _, input := range []string{"", "   \n", "{}\n", "[]"} {
		plan := ConnectClusterResourceModel{
			Name: types.StringValue("cluster-a"),
			Plugins: []ConnectClusterPluginModel{
				{Name: types.StringValue("s3-sink"), Version: types.StringValue("11.1.0")},
			},
			KafkaCluster: &ConnectClusterKafkaModel{KafkaInstanceID: types.StringValue("kf-1")},
			Capacity: &ConnectClusterCapacityModel{
				Type: types.StringValue("provisioned"),
				Provisioned: &ConnectClusterProvisionedCapacityModel{
					WorkerResourceSpec: types.StringValue("TIER1"),
					WorkerCount:        types.Int64Value(1),
				},
			},
			Compute: &ConnectClusterComputeModel{
				Type: types.StringValue("k8s"),
				Kubernetes: &ConnectClusterKubernetesModel{
					ClusterID:      types.StringValue("eks-1"),
					Namespace:      types.StringValue("connect"),
					ServiceAccount: types.StringValue("connect-sa"),
					SchedulingSpec: types.StringValue(input),
				},
			},
		}

		req, diags := ExpandConnectClusterCreate(plan)
		assert.False(t, diags.HasError())
		assert.Nil(t, req.Compute.Kubernetes.SchedulingSpec)
	}
}
