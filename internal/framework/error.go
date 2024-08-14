package framework

import "terraform-provider-automq/client"

func IsNotFoundError(err error) bool {
	condition, ok := err.(*client.ErrorResponse)
	return ok && condition.Code == 404
}
