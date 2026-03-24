package client

import "time"

// Plugin lifecycle states returned by the API.
const (
	PluginStateActive   = "ACTIVE"
	PluginStateDisabled = "DISABLED"
	PluginStateDeleting = "DELETING"
	PluginStateDeleted  = "DELETED"
	PluginStatePending  = "PENDING"
)

// Plugin provider types.
const (
	PluginProviderAutoMQ = "AUTOMQ"
	PluginProviderCustom = "CUSTOM"
)

// ---------------------------------------------------------------------------
// Request params
// ---------------------------------------------------------------------------

type ConnectPluginCreateParam struct {
	Name              string   `json:"name"`
	Version           string   `json:"version"`
	StorageUrl        string   `json:"storageUrl"`
	Types             []string `json:"types"`
	ConnectorClass    string   `json:"connectorClass"`
	Description       *string  `json:"description,omitempty"`
	DocumentationLink *string  `json:"documentationLink,omitempty"`
}

// ---------------------------------------------------------------------------
// Response VOs
// ---------------------------------------------------------------------------

type ConnectPluginVO struct {
	Id                     *string    `json:"id,omitempty"`
	Name                   *string    `json:"name,omitempty"`
	Description            *string    `json:"description,omitempty"`
	DocumentationLink      *string    `json:"documentationLink,omitempty"`
	Types                  []string   `json:"types,omitempty"`
	Provider               *string    `json:"provider,omitempty"`
	StorageUrl             *string    `json:"storageUrl,omitempty"`
	Status                 *string    `json:"status,omitempty"`
	Version                *string    `json:"version,omitempty"`
	ConnectorClass         *string    `json:"connectorClass,omitempty"`
	SourceConnectorClasses []string   `json:"sourceConnectorClasses,omitempty"`
	SinkConnectorClasses   []string   `json:"sinkConnectorClasses,omitempty"`
	CreateTime             *time.Time `json:"createTime,omitempty"`
	UpdateTime             *time.Time `json:"updateTime,omitempty"`
}
