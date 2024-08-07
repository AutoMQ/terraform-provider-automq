package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/wiremock/go-wiremock"
)

const (
	expectedCountOne = int64(1)
)

var createKafkaInstancePath = "/api/v1/instances"
var getKafkaInstancePath = "/api/v1/instances/%s"

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
	createInstanceStub := wiremock.Post(wiremock.
		URLPathEqualTo(createKafkaInstancePath)).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(creatingResponseJson)).WithStatus(http.StatusOK))
	_ = wiremockClient.StubFor(createInstanceStub)

	getInstanceStubWhenStarted := wiremock.Get(wiremock.
		URLPathEqualTo(fmt.Sprintf(getKafkaInstancePath, creatingResponse.InstanceID))).
		WillReturnResponse(wiremock.NewResponse().WithBody(string(availableResponseJson)).WithStatus(http.StatusOK)).
		InScenario("KafkaInstanceState").WhenScenarioStateIs(wiremock.ScenarioStateStarted)
	_ = wiremockClient.StubFor(getInstanceStubWhenStarted)

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
  byoc_host  = "%s"
  token = "123456"
}
resource "automq_kafka_instance" "test" {
  name   = "test"
  description    = "test"
  cloud_provider = "aliyun"
  region         = "cn-hangzhou"
  network_type   = "vpc"
  networks = [{
    zone   = "cn-hangzhou-b"
    subnet = "vsw-bp14v5eikr8wrgoqje7hr"
  }]
  compute_specs = {
    aku = "6"
  }
}
`, mockServerUrl)
}

// Return a json string for a KafkaInstanceResponse with Creating status
func testAccKafkaInstanceResponseInCreating() client.KafkaInstanceResponse {
	createInstanceResponse := client.KafkaInstanceResponse{}
	createInstanceResponse.Status = stateCreating
	createInstanceResponse.DisplayName = "test"
	createInstanceResponse.InstanceID = "test"
	return createInstanceResponse
}

// Return a json string for a KafkaInstanceResponse with Available status
func testAccKafkaInstanceResponseInAvailable() client.KafkaInstanceResponse {
	createInstanceResponse := client.KafkaInstanceResponse{}
	createInstanceResponse.Status = stateAvailable
	createInstanceResponse.DisplayName = "test"
	createInstanceResponse.InstanceID = "test"
	return createInstanceResponse
}

// Return a json string for a KafkaInstanceResponse with Available status
func testAccKafkaInstanceResponseInDeleting() client.KafkaInstanceResponse {
	createInstanceResponse := client.KafkaInstanceResponse{}
	createInstanceResponse.Status = stateDeleting
	createInstanceResponse.DisplayName = "test"
	createInstanceResponse.InstanceID = "test"
	return createInstanceResponse
}
