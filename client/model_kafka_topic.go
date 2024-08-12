package client

// TopicVO struct for TopicVO
type TopicVO struct {
	TopicId   string                 `json:"topicId,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Partition int64                  `json:"partition,omitempty"`
	Configs   map[string]interface{} `json:"configs,omitempty"`
}

// TopicPartitionParam struct for TopicPartitionParam
type TopicPartitionParam struct {
	Partition int64 `json:"partition"`
}

// TopicCreateParam struct for TopicCreateParam
type TopicCreateParam struct {
	Name            string            `json:"name" validate:"regexp=^[a-zA-Z0-9][.a-zA-Z0-9_-]*[a-zA-Z0-9]$"`
	Partition       int64             `json:"partition"`
	CompactStrategy string            `json:"compactStrategy"`
	Configs         []ConfigItemParam `json:"configs,omitempty"`
}

// TopicConfigParam struct for TopicConfigParam
type TopicConfigParam struct {
	Configs []ConfigItemParam `json:"configs"`
}

// TopicApiQuery struct for TopicApiQuery
type TopicApiQuery struct {
	Page     int32  `json:"page,omitempty"`
	Size     int32  `json:"size,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Desc     bool   `json:"desc,omitempty"`
	Keyword  string `json:"keyword,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}
