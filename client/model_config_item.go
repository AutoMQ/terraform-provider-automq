package client

// ConfigItemParam represents a configuration item parameter
type ConfigItemParam struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
	Name  *string `json:"name,omitempty"`
}
