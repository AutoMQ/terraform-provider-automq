//go:build integration

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"terraform-provider-automq/client"
)

const (
	endpointEnv   = "AUTOMQ_ENDPOINT"
	accessKeyEnv  = "AUTOMQ_ACCESS_KEY"
	secretKeyEnv  = "AUTOMQ_SECRET_KEY"
	envIDEnv      = "AUTOMQ_ENV_ID"
	createBodyEnv = "AUTOMQ_KAFKA_CREATE_PAYLOAD"
)

func TestDefaultKafkaInstanceAPI_CreateGetDelete(t *testing.T) {
	endpoint := strings.TrimSpace(os.Getenv(endpointEnv))
	accessKey := strings.TrimSpace(os.Getenv(accessKeyEnv))
	secretKey := strings.TrimSpace(os.Getenv(secretKeyEnv))
	envID := strings.TrimSpace(os.Getenv(envIDEnv))
	createPayload := strings.TrimSpace(os.Getenv(createBodyEnv))

	if endpoint == "" || accessKey == "" || secretKey == "" || envID == "" || createPayload == "" {
		t.Skipf("missing API configuration; ensure %s, %s, %s, %s, %s are set", endpointEnv, accessKeyEnv, secretKeyEnv, envIDEnv, createBodyEnv)
	}

	t.Logf("create payload input: %s", createPayload)

	var create client.InstanceCreateParam
	if err := json.Unmarshal([]byte(createPayload), &create); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", createBodyEnv, err)
	}

	if strings.TrimSpace(create.Name) == "" {
		create.Name = "tf-integration"
	}

	create.Name = ensureNameWithSuffix(create.Name)
	ctx := context.WithValue(context.Background(), client.EnvIdKey, envID)

	cli, err := client.NewClient(ctx, endpoint, client.AuthCredentials{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	})
	if err != nil {
		t.Fatalf("failed to create API client: %v", err)
	}

	api := defaultKafkaInstanceAPI{client: cli}
	summary, err := api.CreateKafkaInstance(ctx, create)
	if err != nil {
		t.Fatalf("CreateKafkaInstance failed: %v", err)
	}
	if summary == nil || summary.InstanceId == nil {
		t.Fatal("CreateKafkaInstance returned nil instance ID")
	}
	logAsJSON(t, "CreateKafkaInstance response", summary)

	instanceID := *summary.InstanceId
	t.Logf("created instance %s with id %s", create.Name, instanceID)

	t.Cleanup(func() {
		delErr := api.DeleteKafkaInstance(ctx, instanceID)
		if delErr != nil {
			var apiErr *client.ErrorResponse
			if errors.As(delErr, &apiErr) && apiErr.Code == 404 {
				return
			}
			t.Logf("DeleteKafkaInstance cleanup error: %v", delErr)
		}
	})

	inst, err := waitForInstanceReady(t, ctx, api, instanceID, 30*time.Minute)
	if err != nil {
		t.Fatalf("waiting for instance %s: %v", instanceID, err)
	}
	logAsJSON(t, "GetKafkaInstance response", inst)
	if inst.Name == nil || *inst.Name != create.Name {
		t.Fatalf("expected instance name %q, got %v", create.Name, inst.Name)
	}

	instByName, err := waitForInstanceByName(t, ctx, api, create.Name, 10*time.Minute)
	if err != nil {
		t.Fatalf("GetKafkaInstanceByName failed: %v", err)
	}
	logAsJSON(t, "GetKafkaInstanceByName response", instByName)
	if instByName.InstanceId == nil || *instByName.InstanceId != instanceID {
		t.Fatalf("GetKafkaInstanceByName returned id %v, want %s", instByName.InstanceId, instanceID)
	}

	if err := api.DeleteKafkaInstance(ctx, instanceID); err != nil {
		t.Fatalf("DeleteKafkaInstance failed: %v", err)
	}
	t.Logf("DeleteKafkaInstance issued for %s", instanceID)

	if err := waitForInstanceDeletion(ctx, api, instanceID, 30*time.Minute); err != nil {
		t.Fatalf("instance %s deletion verification failed: %v", instanceID, err)
	}
}

func ensureNameWithSuffix(base string) string {
	trimmed := strings.TrimRight(base, "-")
	if trimmed == "" {
		trimmed = "tf-integration"
	}
	suffix := time.Now().UTC().Format("20060102-150405")
	name := fmt.Sprintf("%s-%s", trimmed, suffix)
	if len(name) <= 64 {
		return name
	}
	maxBaseLen := 64 - len(suffix) - 1
	if maxBaseLen < 1 {
		return name[len(name)-64:]
	}
	return fmt.Sprintf("%s-%s", trimmed[:maxBaseLen], suffix)
}

func waitForInstanceReady(t *testing.T, ctx context.Context, api defaultKafkaInstanceAPI, instanceID string, timeout time.Duration) (*client.InstanceVO, error) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var (
		lastErr   error
		lastState string
	)

	for {
		inst, err := api.GetKafkaInstance(ctx, instanceID)
		if err != nil {
			lastErr = err
			if shouldStopRetry(err) {
				return nil, err
			}
		} else {
			state := safeState(inst.State)
			if state != lastState {
				t.Logf("instance %s state -> %s", instanceID, state)
				lastState = state
			}

			switch state {
			case "Running":
				return inst, nil
			case "Error":
				return nil, fmt.Errorf("instance %s entered Error state", instanceID)
			case "Deleting", "Deleted":
				return nil, fmt.Errorf("instance %s unexpectedly in %s state", instanceID, state)
			}

			lastErr = fmt.Errorf("instance state %s", state)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for instance %s to become Ready: %w", instanceID, lastErr)
		}

		time.Sleep(15 * time.Second)
	}
}

func waitForInstanceByName(t *testing.T, ctx context.Context, api defaultKafkaInstanceAPI, name string, timeout time.Duration) (*client.InstanceVO, error) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var (
		lastErr   error
		lastState string
	)

	for {
		inst, err := api.GetKafkaInstanceByName(ctx, name)
		if err != nil {
			lastErr = err
			if shouldStopRetry(err) {
				return nil, err
			}
		} else {
			state := safeState(inst.State)
			if state != lastState {
				t.Logf("lookup by name %s state -> %s", name, state)
				lastState = state
			}

			switch state {
			case "Running":
				return inst, nil
			case "Error":
				return nil, fmt.Errorf("instance %s reported Error state", name)
			case "Deleting", "Deleted":
				return nil, fmt.Errorf("instance %s unexpectedly in %s state", name, state)
			}

			lastErr = fmt.Errorf("instance state %s", state)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for instance %s by name: %w", name, lastErr)
		}

		time.Sleep(15 * time.Second)
	}
}

func waitForInstanceDeletion(ctx context.Context, api defaultKafkaInstanceAPI, instanceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		_, err := api.GetKafkaInstance(ctx, instanceID)
		if err != nil {
			var apiErr *client.ErrorResponse
			if errors.As(err, &apiErr) && apiErr.Code == 404 {
				return nil
			}
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for instance %s deletion", instanceID)
		}
		time.Sleep(15 * time.Second)
	}
}

func shouldStopRetry(err error) bool {
	var apiErr *client.ErrorResponse
	if errors.As(err, &apiErr) {
		if apiErr.Code == 400 || apiErr.Code == 403 {
			return true
		}
	}
	return false
}

func logAsJSON(t *testing.T, label string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Logf("%s (marshal error): %v", label, err)
		return
	}
	t.Logf("%s:\n%s", label, string(b))
}

func safeState(statePtr *string) string {
	if statePtr == nil {
		return ""
	}
	return *statePtr
}
