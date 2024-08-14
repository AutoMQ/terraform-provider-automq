package client

// KafkaUserVO struct for KafkaUserVO
type KafkaUserVO struct {
	Name                    string   `json:"name"`
	Password                string   `json:"password"`
	SupportedSaslMechanisms []string `json:"supportedSaslMechanisms,omitempty"`
}

// InstanceUserCreateParam struct for InstanceUserCreateParam
type InstanceUserCreateParam struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// PageNumResultKafkaUserVO struct for PageNumResultKafkaUserVO
type PageNumResultKafkaUserVO struct {
	PageNum   *int32        `json:"pageNum,omitempty"`
	PageSize  *int32        `json:"pageSize,omitempty"`
	Total     *int64        `json:"total,omitempty"`
	List      []KafkaUserVO `json:"list,omitempty"`
	TotalPage *int64        `json:"totalPage,omitempty"`
}
