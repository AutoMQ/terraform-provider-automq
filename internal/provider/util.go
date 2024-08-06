package provider

import (
	"fmt"
	"reflect"
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

type GenericOpenAPIError interface {
	Model() interface{}
}

// createDescriptiveError will convert GenericOpenAPIError error into an error with a more descriptive error message.
// diag.FromErr(createDescriptiveError(err)) should be used instead of diag.FromErr(err) in this project
// since GenericOpenAPIError.Error() returns just HTTP status code and its generic name (i.e., "400 Bad Request")
func createDescriptiveError(err error) error {
	if err == nil {
		return nil
	}
	// At this point it's just status code and its generic name
	errorMessage := err.Error()
	// Add error.detail to the final error message
	if genericOpenAPIError, ok := err.(GenericOpenAPIError); ok {
		failure := genericOpenAPIError.Model()
		reflectedFailure := reflect.ValueOf(&failure).Elem().Elem()
		reflectedFailureValue := reflect.Indirect(reflectedFailure)
		if reflectedFailureValue.IsValid() {
			errs := reflectedFailureValue.FieldByName("Errors")
			kafkaRestOrConnectErr := reflectedFailureValue.FieldByName("Message")
			if errs.Kind() == reflect.Slice && errs.Len() > 0 {
				nest := errs.Index(0)
				detailPtr := nest.FieldByName("Detail")
				if detailPtr.IsValid() {
					errorMessage = fmt.Sprintf("%s: %s", errorMessage, reflect.Indirect(detailPtr))
				}
			} else if kafkaRestOrConnectErr.IsValid() && kafkaRestOrConnectErr.Kind() == reflect.Struct {
				detailPtr := kafkaRestOrConnectErr.FieldByName("value")
				if detailPtr.IsValid() {
					errorMessage = fmt.Sprintf("%s: %s", errorMessage, reflect.Indirect(detailPtr))
				}
			} else if kafkaRestOrConnectErr.IsValid() && kafkaRestOrConnectErr.Kind() == reflect.Pointer {
				errorMessage = fmt.Sprintf("%s: %s", errorMessage, reflect.Indirect(kafkaRestOrConnectErr))
			}
		}
	}
	return fmt.Errorf(errorMessage)
}
