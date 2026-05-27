package provider

import (
	"context"
	"errors"
	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testStringPtr(s string) *string {
	return &s
}

type stubKafkaInstanceAPI struct {
	instance         *client.InstanceVO
	endpoints        []client.InstanceAccessInfoVO
	getInstanceErr   error
	getEndpointsErr  error
	getEndpointsCall int
}

func (s *stubKafkaInstanceAPI) CreateKafkaInstance(context.Context, client.InstanceCreateParam) (*client.InstanceSummaryVO, error) {
	return nil, errors.New("unexpected CreateKafkaInstance call")
}

func (s *stubKafkaInstanceAPI) GetKafkaInstance(context.Context, string) (*client.InstanceVO, error) {
	return s.instance, s.getInstanceErr
}

func (s *stubKafkaInstanceAPI) GetKafkaInstanceByName(context.Context, string) (*client.InstanceVO, error) {
	return nil, errors.New("unexpected GetKafkaInstanceByName call")
}

func (s *stubKafkaInstanceAPI) DeleteKafkaInstance(context.Context, string) error {
	return errors.New("unexpected DeleteKafkaInstance call")
}

func (s *stubKafkaInstanceAPI) UpdateKafkaInstance(context.Context, string, client.InstanceUpdateParam) error {
	return errors.New("unexpected UpdateKafkaInstance call")
}

func (s *stubKafkaInstanceAPI) GetInstanceEndpoints(context.Context, string) ([]client.InstanceAccessInfoVO, error) {
	s.getEndpointsCall++
	return s.endpoints, s.getEndpointsErr
}

func TestRefreshKafkaInstanceState_FetchesEndpointsOnlyForRunningInstance(t *testing.T) {
	t.Run("running instance refreshes endpoints", func(t *testing.T) {
		api := &stubKafkaInstanceAPI{
			instance: &client.InstanceVO{
				InstanceId: testStringPtr("inst-1"),
				Name:       testStringPtr("test"),
				State:      testStringPtr(models.StateRunning),
			},
			endpoints: []client.InstanceAccessInfoVO{
				{
					DisplayName:      testStringPtr("private"),
					NetworkType:      testStringPtr("private"),
					Protocol:         testStringPtr("SASL_PLAINTEXT"),
					BootstrapServers: testStringPtr("broker:9092"),
				},
			},
		}
		resource := &KafkaInstanceResource{api: api}
		state := models.KafkaInstanceResourceModel{}

		_, found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-1", &state)
		if !found {
			t.Fatalf("expected instance to be found")
		}
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if api.getEndpointsCall != 1 {
			t.Fatalf("expected exactly one endpoints call, got %d", api.getEndpointsCall)
		}
	})

	t.Run("non-running instance skips endpoints", func(t *testing.T) {
		api := &stubKafkaInstanceAPI{
			instance: &client.InstanceVO{
				InstanceId: testStringPtr("inst-2"),
				Name:       testStringPtr("test"),
				State:      testStringPtr(models.StateCreating),
			},
		}
		resource := &KafkaInstanceResource{api: api}
		state := models.KafkaInstanceResourceModel{}

		_, found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-2", &state)
		if !found {
			t.Fatalf("expected instance to be found")
		}
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if api.getEndpointsCall != 0 {
			t.Fatalf("expected no endpoints call, got %d", api.getEndpointsCall)
		}
		if !state.Endpoints.IsNull() && !state.Endpoints.IsUnknown() && len(state.Endpoints.Elements()) != 0 {
			t.Fatalf("expected endpoints to remain unset for non-running instance")
		}
	})
}

func TestRefreshKafkaInstanceState_NotFound(t *testing.T) {
	resource := &KafkaInstanceResource{
		api: &stubKafkaInstanceAPI{
			getInstanceErr: &client.ErrorResponse{Code: 404},
		},
	}
	state := models.KafkaInstanceResourceModel{
		InstanceID: types.StringValue("inst-missing"),
	}

	_, found, diags := refreshKafkaInstanceState(context.Background(), resource, "inst-missing", &state)
	if found {
		t.Fatalf("expected instance to be treated as not found")
	}
	if diags.HasError() {
		t.Fatalf("expected no diagnostics for not found, got %v", diags)
	}
}
