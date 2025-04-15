package framework

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func WaitForKafkaClusterState(ctx context.Context, c *client.Client, clusterId, pendingState, targetState string, timeout time.Duration, refreshFunc retry.StateRefreshFunc) error {
	delay, pollInterval := 20*time.Second, 10*time.Second
	stateConf := &retry.StateChangeConf{
		Pending:      []string{pendingState},
		Target:       []string{targetState},
		Refresh:      refreshFunc,
		Timeout:      timeout,
		Delay:        delay,
		PollInterval: pollInterval,
	}

	tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q status to become %q", clusterId, targetState))
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func WaitForKafkaClusterToProvision(ctx context.Context, c *client.Client, clusterId, pendingState string, timeout time.Duration) error {
	return WaitForKafkaClusterState(ctx, c, clusterId, pendingState, models.StateRunning, timeout, KafkaClusterStatus(ctx, c, clusterId, models.StateRunning))
}

func WaitForKafkaClusterToDeleted(ctx context.Context, c *client.Client, clusterId string, timeout time.Duration) error {
	return WaitForKafkaClusterState(ctx, c, clusterId, models.StateDeleting, models.StateNotFound, timeout, KafkaClusterStatus(ctx, c, clusterId, models.StateNotFound))
}

func KafkaClusterStatus(ctx context.Context, c *client.Client, clusterId string, targetState string) retry.StateRefreshFunc {
	return func() (result interface{}, s string, err error) {
		cluster, err := c.GetKafkaInstance(ctx, clusterId)
		if err != nil {
			if IsNotFoundError(err) {
				return &client.KafkaInstanceResponse{}, models.StateNotFound, nil
			}
			tflog.Warn(ctx, fmt.Sprintf("Error reading Kafka Cluster %q: %s", clusterId, err))
			return nil, models.StateUnknown, err
		}
		if cluster == nil {
			return nil, models.StateUnknown, fmt.Errorf("Kafka Cluster %q not found", clusterId)
		}

		tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q status to become %q: current status is %q", clusterId, targetState, *cluster.State))
		if *cluster.State == models.StateError {
			return nil, models.StateError, fmt.Errorf("Kafka Cluster %q status is %q", clusterId, models.StateError)
		}
		return cluster, *cluster.State, nil
	}
}
