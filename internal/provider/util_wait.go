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

func waitForKafkaClusterToProvision(ctx context.Context, c *client.Client, clusterId, cloudProvider string) error {
	delay, pollInterval := getDelayAndPollInterval(5*time.Second, 1*time.Minute, false)
	stateConf := &retry.StateChangeConf{
		Pending:      []string{stateCreating},
		Target:       []string{stateAvailable},
		Refresh:      kafkaClusterProvisionStatus(ctx, c, clusterId),
		Timeout:      getTimeoutFor(cloudProvider),
		Delay:        delay,
		PollInterval: pollInterval,
	}

	tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q provisioning status to become %q", clusterId, stateAvailable))
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func waitForKafkaClusterToDeleted(ctx context.Context, c *client.Client, clusterId, cloudProvider string) error {
	delay, pollInterval := getDelayAndPollInterval(5*time.Second, 1*time.Minute, false)
	stateConf := &retry.StateChangeConf{
		Pending:      []string{stateDeleting},
		Target:       []string{stateNotFound},
		Refresh:      kafkaClusterDeletedStatus(ctx, c, clusterId),
		Timeout:      getTimeoutFor(cloudProvider),
		Delay:        delay,
		PollInterval: pollInterval,
	}

	tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q provisioning status to become %q", clusterId, stateAvailable))
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return err
	}
	return nil
}

func kafkaClusterProvisionStatus(ctx context.Context, c *client.Client, clusterId string) retry.StateRefreshFunc {
	return func() (result interface{}, s string, err error) {
		cluster, err := c.GetKafkaInstance(clusterId)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Error reading Kafka Cluster %q: %s", clusterId, createDescriptiveError(err)))
			return nil, stateUnknown, err
		}

		tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q provisioning status to become %q: current status is %q", clusterId, stateAvailable, cluster.Status))
		if cluster.Status == stateCreating || cluster.Status == stateAvailable {
			return cluster, cluster.Status, nil
		} else if cluster.Status == stateError {
			return nil, stateError, fmt.Errorf("kafka Cluster %q provisioning status is %q", clusterId, stateError)
		}
		// Kafka Cluster is in an unexpected state
		return nil, stateUnexpected, fmt.Errorf("kafka Cluster %q is an unexpected state %q", clusterId, cluster.Status)
	}
}

func kafkaClusterDeletedStatus(ctx context.Context, c *client.Client, clusterId string) retry.StateRefreshFunc {
	return func() (result interface{}, s string, err error) {
		cluster, err := c.GetKafkaInstance(clusterId)
		if err != nil {
			tflog.Warn(ctx, fmt.Sprintf("Error reading Kafka Cluster %q: %s", clusterId, createDescriptiveError(err)))
			return nil, stateUnknown, err
		}
		if cluster == nil {
			return nil, stateNotFound, nil
		}
		tflog.Debug(ctx, fmt.Sprintf("Waiting for Kafka Cluster %q provisioning status to become %q: current status is %q", clusterId, stateDeleting, cluster.Status))
		if cluster.Status == stateError {
			return nil, stateError, fmt.Errorf("kafka Cluster %q provisioning status is %q", clusterId, stateError)
		} else {
			return cluster, cluster.Status, nil
		}
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
