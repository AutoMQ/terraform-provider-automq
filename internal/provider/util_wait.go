package provider

import (
	"context"
	"fmt"
	"terraform-provider-automq/client"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	stateUp                        = "UP"
	stateCreated                   = "CREATED"
	acceptanceTestModeWaitTime     = 1 * time.Second
	acceptanceTestModePollInterval = 1 * time.Second
)

func waitForKafkaClusterState(ctx context.Context, c *client.Client, clusterId, cloudProvider, pendingState, targetState string, refreshFunc retry.StateRefreshFunc) error {
	delay, pollInterval := getDelayAndPollInterval(5*time.Second, 1*time.Minute, false)
	stateConf := &retry.StateChangeConf{
		Pending:      []string{pendingState},
		Target:       []string{targetState},
		Refresh:      refreshFunc,
		Timeout:      getTimeoutFor(cloudProvider),
		Delay:        delay,
		PollInterval: pollInterval,
	}

	tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q status to become %q", clusterId, targetState))
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func waitForKafkaClusterToProvision(ctx context.Context, c *client.Client, clusterId, cloudProvider, pendingState string) error {
	return waitForKafkaClusterState(ctx, c, clusterId, cloudProvider, pendingState, stateAvailable, kafkaClusterStatus(ctx, c, clusterId, stateAvailable))
}

func waitForKafkaClusterToDeleted(ctx context.Context, c *client.Client, clusterId, cloudProvider string) error {
	return waitForKafkaClusterState(ctx, c, clusterId, cloudProvider, stateDeleting, stateNotFound, kafkaClusterStatus(ctx, c, clusterId, stateNotFound))
}

func kafkaClusterStatus(ctx context.Context, c *client.Client, clusterId string, targetState string) retry.StateRefreshFunc {
	return func() (result interface{}, s string, err error) {
		cluster, err := c.GetKafkaInstance(clusterId)
		if err != nil {
			if isNotFoundError(err) && targetState == stateNotFound {
				return &client.KafkaInstanceResponse{}, stateNotFound, nil
			}
			tflog.Warn(ctx, fmt.Sprintf("Error reading Kafka Cluster %q: %s", clusterId, err))
			return nil, stateUnknown, err
		}
		if cluster == nil {
			return nil, stateUnknown, fmt.Errorf("Kafka Cluster %q not found", cluster.InstanceID)
		}

		tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q status to become %q: current status is %q", clusterId, targetState, cluster.Status))
		if cluster.Status == stateError {
			return nil, stateError, fmt.Errorf("Kafka Cluster %q status is %q", clusterId, stateError)
		}
		return cluster, cluster.Status, nil
	}
}

// If `isAcceptanceTestMode` is true, default wait time and poll interval are returned
// If `isAcceptanceTestMode` is false, customized wait time and poll interval are returned
func getDelayAndPollInterval(delayNormal, pollIntervalNormal time.Duration, isAcceptanceTestMode bool) (time.Duration, time.Duration) {
	if isAcceptanceTestMode {
		return acceptanceTestModeWaitTime, acceptanceTestModePollInterval
	}
	return delayNormal, pollIntervalNormal
}
