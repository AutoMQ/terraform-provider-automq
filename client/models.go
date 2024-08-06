package client

import "time"

type KafkaInstanceRequest struct {
	DisplayName string                        `json:"displayName"`
	Description string                        `json:"description"`
	Provider    string                        `json:"provider"`
	Region      string                        `json:"region"`
	Spec        KafkaInstanceRequestSpec      `json:"spec"`
	Networks    []KafkaInstanceRequestNetwork `json:"networks"`
}

type KafkaInstanceRequestSpec struct {
	Template    string                          `json:"template"`
	PaymentPlan KafkaInstanceRequestPaymentPlan `json:"paymentPlan"`
	Values      []KafkaInstanceRequestValues    `json:"values"`
}

type KafkaInstanceRequestPaymentPlan struct {
	PaymentType string `json:"paymentType"`
	Period      int    `json:"period"`
	Unit        string `json:"unit"`
}

type KafkaInstanceRequestValues struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KafkaInstanceRequestNetwork struct {
	Zone   string `json:"zone"`
	Subnet string `json:"subnet"`
}

type KafkaInstanceResponse struct {
	InstanceID  string    `json:"instanceId"`
	GmtCreate   time.Time `json:"gmtCreate"`
	GmtModified time.Time `json:"gmtModified"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Provider    string    `json:"provider"`
	Region      string    `json:"region"`
	Spec        struct {
		SpecID      string `json:"specId"`
		DisplayName string `json:"displayName"`
		PaymentPlan struct {
			PaymentType string `json:"paymentType"`
			Unit        string `json:"unit"`
			Period      int    `json:"period"`
		} `json:"paymentPlan"`
		Template string `json:"template"`
		Version  string `json:"version"`
		Values   []struct {
			Key          string `json:"key"`
			Name         string `json:"name"`
			Value        int    `json:"value"`
			DisplayValue string `json:"displayValue"`
		} `json:"values"`
	} `json:"spec"`
	Networks []struct {
		Zone    string `json:"zone"`
		Subnets []struct {
			Subnet     string `json:"subnet"`
			SubnetName string `json:"subnetName"`
		} `json:"subnets"`
	} `json:"networks"`
	Metrics      []interface{} `json:"metrics"`
	AclSupported bool          `json:"aclSupported"`
	AclEnabled   bool          `json:"aclEnabled"`
}
