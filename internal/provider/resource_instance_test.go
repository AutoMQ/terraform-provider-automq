package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"terraform-provider-automq/client"
	"terraform-provider-automq/internal/models"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/wiremock/go-wiremock"
)

const (
	expectedCountOne = int64(1)
)

var createKafkaInstancePath = "/api/v1/instances"
var getKafkaInstancePath = "/api/v1/instances/%s"
var getKafkaInstanceEndpointPath = "/api/v1/instances/%s/endpoints"
var getKafkaIntegrationPath = "/api/v1/instances/%s/integrations"

func TestAccKafkaInstanceResource(t *testing.T) {
	ctx := context.Background()

	wiremockContainer, err := setupWiremock(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// nolint:errcheck
	defer wiremockContainer.Terminate(ctx)

	mockAutoMQTestServerUrl := wiremockContainer.URI
	wiremockClient := wiremock.NewClient(mockAutoMQTestServerUrl)

	creatingResponse := testAccKafkaInstanceResponseInCreating()
	creatingResponseJson, err := json.Marshal(creatingResponse)
	if err != nil {
		t.Fatal(err)
	}

	availableResponse := testAccKafkaInstanceResponseInAvailable()
	availableResponseJson, err := json.Marshal(availableResponse)
	if err != nil {
		t.Fatal(err)
	}
	deletingResponse := testAccKafkaInstanceResponseInDeleting()
	deletingResponseJson, err := json.Marshal(deletingResponse)
	if err != nil {
		t.Fatal(err)
	}
	IntegrationResponse := newInstanceIntegrationResponse()
	IntegrationResponseJson, err := json.Marshal(IntegrationResponse)
	if err != nil {
		t.Fatal(err)
	}

	EndpointsResponse := newInstanceEndpointsResponse()
	EndpointsResponseJson, err := json.Marshal(EndpointsResponse)
	if err != nil {
		t.Fatal(err)
	}
	createInstanceStub := wiremock.Post(wiremock.
		URLPathEqualTo(createKafkaInstancePath)).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(creatingResponseJson)).WithStatus(http.StatusOK))
	_ = wiremockClient.StubFor(createInstanceStub)

	getInstanceStubWhenStarted := wiremock.Get(wiremock.
		URLPathEqualTo(fmt.Sprintf(getKafkaInstancePath, creatingResponse.InstanceID))).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(availableResponseJson)).WithStatus(http.StatusOK)).
		InScenario("KafkaInstanceState").WhenScenarioStateIs(wiremock.ScenarioStateStarted)
	_ = wiremockClient.StubFor(getInstanceStubWhenStarted)

	getKafkaInstanceEndpointStub := wiremock.Get(wiremock.
		URLPathEqualTo(fmt.Sprintf(getKafkaInstanceEndpointPath, creatingResponse.InstanceID))).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(EndpointsResponseJson)).WithStatus(http.StatusOK))
	_ = wiremockClient.StubFor(getKafkaInstanceEndpointStub)

	getInstanceIntegrationStubWhenStarted := wiremock.Get(wiremock.
		URLPathEqualTo(fmt.Sprintf(getKafkaIntegrationPath, creatingResponse.InstanceID))).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(IntegrationResponseJson)).WithStatus(http.StatusOK))
	_ = wiremockClient.StubFor(getInstanceIntegrationStubWhenStarted)

	deleteInstanceStub := wiremock.Delete(wiremock.
		URLPathEqualTo(fmt.Sprintf(getKafkaInstancePath, creatingResponse.InstanceID))).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(deletingResponseJson)).WithStatus(http.StatusNoContent)).
		InScenario("KafkaInstanceState").WillSetStateTo("Deleted")
	_ = wiremockClient.StubFor(deleteInstanceStub)

	getInstanceStubWhenDeleted := wiremock.Get(wiremock.
		URLPathEqualTo(fmt.Sprintf(getKafkaInstancePath, creatingResponse.InstanceID))).
		WillReturnResponse(wiremock.NewResponse().WithStatus(http.StatusNotFound)).
		InScenario("KafkaInstanceState").WhenScenarioStateIs("Deleted")
	_ = wiremockClient.StubFor(getInstanceStubWhenDeleted)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKafkaInstanceResourceConfig(mockAutoMQTestServerUrl),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("automq_kafka_instance.test", "name", "test"),
				),
			},
		},
	})

	checkStubCount(t, wiremockClient, createInstanceStub, fmt.Sprintf("POST %s", createKafkaInstancePath), expectedCountOne)
	checkStubCount(t, wiremockClient, deleteInstanceStub, fmt.Sprintf("DELETE %s", getKafkaInstancePath), expectedCountOne)
}

func testAccKafkaInstanceResourceConfig(mockServerUrl string) string {
	return fmt.Sprintf(`
provider "automq" {
  automq_byoc_endpoint  = "%s"
  automq_byoc_access_key_id = "VLaUIeNYndeOAXjaol32o4UAHvX8A7VE"
  automq_byoc_secret_key = "CHlRi0hOIA8pAnzW"
}
resource "automq_kafka_instance" "test" {
  environment_id = "env-1"
  name   = "test"
  description    = "test"
  cloud_provider = "aliyun"
  region         = "cn-hangzhou"
  networks = [{
    zone   = "cn-hangzhou-b"
    subnets = ["vsw-bp14v5eikr8wrgoqje7hr"]
  }]
  compute_specs = {
    aku = "6"
  }
  integrations = ["integration-1"]
}
`, mockServerUrl)
}

// Return a json string for a KafkaInstanceResponse with Creating status
func testAccKafkaInstanceResponseInCreating() client.KafkaInstanceResponse {
	instanceResponse := newInstanceResponse()

	instanceResponse.Status = models.StateCreating
	instanceResponse.GmtCreate = time.Now()
	instanceResponse.GmtModified = time.Now()
	return instanceResponse
}

// Return a json string for a KafkaInstanceResponse with Available status
func testAccKafkaInstanceResponseInAvailable() client.KafkaInstanceResponse {
	instanceResponse := newInstanceResponse()

	instanceResponse.Status = models.StateAvailable
	instanceResponse.GmtModified = time.Now()
	return instanceResponse
}

// Return a json string for a KafkaInstanceResponse with Available status
func testAccKafkaInstanceResponseInDeleting() client.KafkaInstanceResponse {
	instanceResponse := newInstanceResponse()

	instanceResponse.Status = models.StateDeleting
	instanceResponse.GmtModified = time.Now()
	return instanceResponse
}

func newInstanceResponse() client.KafkaInstanceResponse {
	instanceResponse := client.KafkaInstanceResponse{}
	instanceResponse.InstanceID = "kf-cakz90r71mspc7vy"
	instanceResponse.DisplayName = "test"
	instanceResponse.Description = "test"
	instanceResponse.Provider = "aliyun"
	instanceResponse.Region = "cn-hangzhou"
	instanceResponse.Spec.Version = "1.2.0"
	instanceResponse.Spec.PaymentPlan.PaymentType = "ON_DEMAND"
	instanceResponse.Spec.PaymentPlan.Period = 1
	instanceResponse.Spec.PaymentPlan.Unit = "MONTH"
	instanceResponse.Spec.Values = []client.Value{{Key: "aku", Value: 6}}
	instanceResponse.Networks = []client.Network{{Zone: "cn-hangzhou-b", Subnets: []client.Subnet{{Subnet: "vsw-bp14v5eikr8wrgoqje7hr"}}}}
	return instanceResponse
}

func newInstanceIntegrationResponse() client.PageNumResultIntegrationVO {
	return client.PageNumResultIntegrationVO{
		List: []client.IntegrationVO{
			{
				Type:     "cloudwatch",
				Code:     "integration-1",
				Name:     "cloudwatch",
				EndPoint: nil,
				Config: map[string]interface{}{
					"namespace": "example",
				},
				GmtCreate:   time.Now(),
				GmtModified: time.Now(),
			},
		},
	}
}

func newInstanceEndpointsResponse() client.PageNumResultInstanceAccessInfoVO {
	return client.PageNumResultInstanceAccessInfoVO{
		List: []client.InstanceAccessInfoVO{
			{
				BootstrapServers: "kafka-1:9092",
				DisplayName:      "kafka-1",
				NetworkType:      "VPC",
				Protocol:         "PLAINTEXT",
				Mechanisms:       "PLAIN",
			},
		},
	}
}
