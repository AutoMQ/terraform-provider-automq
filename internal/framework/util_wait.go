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
		if unexpected, ok := err.(*retry.UnexpectedStateError); ok && unexpected.LastError == nil {
			return fmt.Errorf("Kafka Cluster %q entered unexpected state %q while waiting for state %q", clusterId, unexpected.State, targetState)
		}
		if timeout, ok := err.(*retry.TimeoutError); ok && timeout.LastError == nil {
			lastState := timeout.LastState
			if lastState == "" {
				lastState = models.StateUnknown
			}
			return fmt.Errorf("Kafka Cluster %q did not reach state %q (last state %q, timeout %s)", clusterId, targetState, lastState, timeout.Timeout)
		}
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
	var lastState string
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

		currentState := derefState(cluster.State)
		if currentState != lastState {
			message := fmt.Sprintf("Kafka Cluster %q state -> %s", clusterId, currentState)
			if currentState == models.StateError {
				tflog.Error(ctx, message)
			} else {
				tflog.Info(ctx, message)
			}
			lastState = currentState
		}

		if currentState == models.StateError {
			return nil, models.StateError, fmt.Errorf("Kafka Cluster %q status is %q", clusterId, models.StateError)
		}
		return cluster, currentState, nil
	}
}

func derefState(statePtr *string) string {
	if statePtr == nil {
		return models.StateUnknown
	}
	return *statePtr
}
