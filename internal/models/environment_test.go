package models

import (
	"testing"
	"time"

	"terraform-provider-automq/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandEnvironment(t *testing.T) {
	model := EnvironmentResourceModel{
		Name:          types.StringValue("production"),
		Description:   types.StringValue("Production environment"),
		CloudProvider: types.StringValue("aws"),
		Region:        types.StringValue("us-east-1"),
		Scope:         types.StringValue("123456789012"),
		Product:       types.StringValue("kafka"),
	}

	create := ExpandEnvironmentCreate(model)
	assert.Equal(t, "production", create.Name)
	assert.Equal(t, "Production environment", *create.Description)
	assert.Equal(t, "aws", create.CloudProvider)
	assert.Equal(t, "us-east-1", create.Region)
	assert.Equal(t, "123456789012", create.Scope)
	assert.Equal(t, "kafka", create.Product)

	update := ExpandEnvironmentUpdate(model)
	assert.Equal(t, "production", update.Name)
	assert.Equal(t, "Production environment", *update.Description)
}

func TestFlattenEnvironmentPreservesProduct(t *testing.T) {
	createdAt := time.Date(2026, time.July, 28, 1, 2, 3, 0, time.UTC)
	description := "Production environment"
	opsBucket := "automq-ops"
	model := EnvironmentResourceModel{Product: types.StringValue("kafka")}

	FlattenEnvironment(&client.EnvironmentVO{
		EnvID:         "env-123",
		CreatedAt:     &createdAt,
		Name:          "production",
		Description:   &description,
		CloudProvider: "aws",
		OpsBucket:     &opsBucket,
		Region:        "us-east-1",
		Scope:         "123456789012",
		State:         "Pending",
		Stateless:     false,
	}, &model)

	assert.Equal(t, "env-123", model.ID.ValueString())
	assert.Equal(t, "production", model.Name.ValueString())
	assert.Equal(t, "aws", model.CloudProvider.ValueString())
	assert.Equal(t, "Pending", model.State.ValueString())
	assert.Equal(t, "kafka", model.Product.ValueString())
	assert.Equal(t, "2026-07-28T01:02:03Z", model.CreatedAt.ValueString())
}
