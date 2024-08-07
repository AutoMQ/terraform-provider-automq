package provider

import (
	"terraform-provider-automq/client"
	"time"
)

const (
	Aliyun = "aliyun"
	AWS    = "aws"
)

func getTimeoutFor(cloudProvider string) time.Duration {
	if cloudProvider == Aliyun {
		return 20 * time.Minute
	} else {
		return 30 * time.Minute
	}
}

func isNotFoundError(err error) bool {
	condition, ok := err.(*client.ErrorResponse)
	return ok && condition.Code == 404
}

type GenericOpenAPIError interface {
	Model() interface{}
}
