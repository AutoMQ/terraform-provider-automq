package client

import "time"

type KafkaInstanceRequest struct {
	DisplayName  string                        `json:"displayName"`
	Description  string                        `json:"description"`
	Provider     string                        `json:"provider"`
	Region       string                        `json:"region"`
	Spec         KafkaInstanceRequestSpec      `json:"spec"`
	Networks     []KafkaInstanceRequestNetwork `json:"networks"`
	AclEnabled   bool                          `json:"aclEnabled"`
	Integrations []string                      `json:"integrations"`
}

type InstanceBasicParam struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type InstanceVersionUpgradeParam struct {
	Version string `json:"version"`
}

type InstanceConfigParam struct {
	Configs []KafkaInstanceRequestValues `json:"configs"`
}

type SpecificationUpdateParam struct {
	Values []KafkaInstanceRequestValues `json:"values"`
}

type KafkaInstanceRequestSpec struct {
	Version     string                          `json:"version"`
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
	InstanceID   string        `json:"instanceId"`
	GmtCreate    time.Time     `json:"gmtCreate"`
	GmtModified  time.Time     `json:"gmtModified"`
	DisplayName  string        `json:"displayName"`
	Description  string        `json:"description"`
	Status       string        `json:"status"`
	Provider     string        `json:"provider"`
	Region       string        `json:"region"`
	Spec         Spec          `json:"spec"`
	Networks     []Network     `json:"networks"`
	Metrics      []interface{} `json:"metrics"`
	AclSupported bool          `json:"aclSupported"`
	AclEnabled   bool          `json:"aclEnabled"`
}

type Spec struct {
	SpecID      string      `json:"specId"`
	DisplayName string      `json:"displayName"`
	PaymentPlan PaymentPlan `json:"paymentPlan"`
	Template    string      `json:"template"`
	Version     string      `json:"version"`
	Values      []Value     `json:"values"`
}

type PaymentPlan struct {
	PaymentType string `json:"paymentType"`
	Unit        string `json:"unit"`
	Period      int    `json:"period"`
}

type Value struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Value        int    `json:"value"`
	DisplayValue string `json:"displayValue"`
}

type Network struct {
	Zone    string   `json:"zone"`
	Subnets []Subnet `json:"subnets"`
}

type Subnet struct {
	Subnet     string `json:"subnet"`
	SubnetName string `json:"subnetName"`
}

type KafkaInstanceResponseList struct {
	PageNum   int                     `json:"pageNum"`
	PageSize  int                     `json:"pageSize"`
	Total     int                     `json:"total"`
	List      []KafkaInstanceResponse `json:"list"`
	TotalPage int                     `json:"totalPage"`
}

type Metric struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Value       int    `json:"value"`
}
