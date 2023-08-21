package main

import (
	"gadu/shared/m/v2/infra"
	"time"
)

type BeeCluster struct {
	ClusterId          uint16 `gorm:"primaryKey"json:"clusterId,omitempty"`
	ClusterName        string `json:"clusterName,omitempty"`
	ClusterDescription string `json:"clusterDescription,omitempty"`
	AutoBind           bool   `json:"autoBind"`
}

type BeeAttribute struct {
	ClusterId         uint16 `gorm:"primaryKey" json:"clusterId,omitempty"`
	ClusterClient     bool   `gorm:"primaryKey" json:"clusterClient,omitempty"`
	AttributeId       uint16 `gorm:"primaryKey" json:"attributeId,omitempty"`
	AttributeMfcode   uint16 `gorm:"primaryKey" json:"attributeMfcode,omitempty"`
	AttributeName     string `json:"attributeName,omitempty"`
	AttributeType     string `json:"attributeType,omitempty"`
	AttributeAccess   string `json:"attributeAccess,omitempty"`
	AttributeDefault  string `json:"attributeDefault,omitempty"`
	AttributeRequired bool   `json:"attributeRequired,omitempty"`
	AttributeShowas   string `json:"attributeShowas,omitempty"`
	AttributeRange0   uint64 `json:"attributeRange0,omitempty"`
	AttributeRange1   uint64 `json:"attributeRange1,omitempty"`
	Behaviour         string `gorm:"->;type:text GENERATED ALWAYS AS (cluster_id || ':' || cluster_client) STORED;" json:"behaviour,omitempty"`
	Property          string `gorm:"->;type:text GENERATED ALWAYS AS (attribute_name) STORED;" json:"property,omitempty"`
	MinInterval       uint16
	MaxInterval       uint16
	ReportableChange  uint16
}

type BeeAttributeValue struct {
	ClusterId           uint16 `gorm:"primaryKey" json:"clusterId,omitempty"`
	ClusterClient       bool   `gorm:"primaryKey" json:"clusterClient,omitempty"`
	AttributeId         uint16 `gorm:"primaryKey" json:"attributeId,omitempty"`
	AttributeMfcode     uint16 `gorm:"primaryKey" json:"attributeMfcode,omitempty"`
	AttributeValueValue uint64 `gorm:"primaryKey" json:"attributeValueValue,omitempty"`
	AttributeValueName  string `json:"attributeValueName,omitempty"`
}

type BeeCommand struct {
	ClusterId        uint16 `gorm:"primaryKey" json:"clusterId,omitempty"`
	ClusterClient    bool   `gorm:"primaryKey" json:"clusterClient,omitempty"`
	CommandId        uint8  `gorm:"primaryKey" json:"commandId,omitempty"`
	CommandName      string `json:"commandName" json:"commandName,omitempty"`
	CommandDirection string `json:"commandDirection" json:"commandDirection,omitempty"`
	CommandMfccode   uint16 `json:"commandMfccode" json:"commandMfccode,omitempty"`
	CommandRequired  bool   `json:"commandRequired,omitempty"`
	CommandResponse  uint16 `json:"commandResponse,omitempty"`
	Behaviour        string `gorm:"->;type:text GENERATED ALWAYS AS (cluster_id || ':' || cluster_client) STORED;"json:"behaviour,omitempty"`
	Command          string `gorm:"->;type:text GENERATED ALWAYS AS (command_name) STORED;"json:"command,omitempty"`
}

type BeeCommandAttribute struct {
	ClusterId                uint16 `gorm:"primaryKey" json:"clusterId,omitempty"`
	ClusterClient            bool   `gorm:"primaryKey" json:"clusterClient,omitempty"`
	CommandId                uint8  `gorm:"primaryKey" json:"commandId,omitempty"`
	CommandAttributeId       uint16 `gorm:"primaryKey" json:"commandAttributeId,omitempty"`
	CommandAttributeType     string `json:"commandAttributeType,omitempty"`
	CommandAttributeName     string `json:"commandAttributeName,omitempty"`
	CommandAttributeDefault  uint16 `json:"commandAttributeDefault,omitempty"`
	CommandAttributeShowas   string `json:"commandAttributeShowas,omitempty"`
	CommandAttributeRequired bool   `json:"commandAttributeRequired,omitempty"`
}

type BeeNode struct {
	IEEEAddress    string    `gorm:"primaryKey" json:"IEEEAddress,omitempty"`
	NetworkAddress uint16    `json:"networkAddress,omitempty"`
	LogicalType    uint8     `json:"logicalType,omitempty"`
	LQI            uint8     `json:"LQI,omitempty"`
	Depth          uint8     `json:"depth,omitempty"`
	LastDiscovered time.Time `json:"lastDiscovered"`
	LastReceived   time.Time `json:"lastReceived"`
	Explored       bool      `json:"explored,omitempty"`
}

type BeeNodeEndpoint struct {
	IEEEAddress      string                   `gorm:"primaryKey" json:"IEEEAddress,omitempty"`
	EndpointId       uint16                   `gorm:"primaryKey" json:"endpointId,omitempty"`
	Name             string                   `json:"name,omitempty"`
	Room             string                   `json:"room,omitempty"`
	ProfileId        uint16                   `json:"profileId,omitempty"`
	DeviceId         uint16                   `json:"deviceId,omitempty"`
	DeviceVersion    uint8                    `json:"deviceVersion,omitempty"`
	Explored         bool                     `json:"explored,omitempty"`
	Enabled          bool                     `json:"enabled,omitempty"`
	ManufacturerName string                   `json:"manufacturerName,omitempty"`
	ModelIdentifier  string                   `json:"modelIdentifier,omitempty"`
	Channel          string                   `gorm:"->;type:text GENERATED ALWAYS AS (ieee_address || ':' || endpoint_id) STORED;"json:"channel,omitempty"`
	Description      string                   `gorm:"->;type:text GENERATED ALWAYS AS (manufacturer_name || ':' || model_identifier) STORED;"json:"description,omitempty"`
	Behaviours       []BeeNodeEndpointCluster `gorm:"foreignKey:ieee_address,endpoint_id"json:"behaviours,omitempty"`
}

type BeeNodeEndpointCluster struct {
	IEEEAddress   string                `gorm:"primaryKey" json:"IEEEAddress,omitempty"`
	EndpointId    uint16                `gorm:"primaryKey" json:"endpointId,omitempty"`
	ClusterId     uint16                `gorm:"primaryKey" json:"clusterId,omitempty"`
	ClusterClient bool                  `gorm:"primaryKey" json:"clusterClient,omitempty"`
	Channel       string                `gorm:"->;type:text GENERATED ALWAYS AS (ieee_address || ':' || endpoint_id) STORED;"json:"channel" json:"channel,omitempty"`
	Behaviour     string                `json:"behaviour,omitempty"`
	Explored      bool                  `json:"explored,omitempty"`
	Enabled       bool                  `json:"enabled,omitempty"`
	Properties    []infra.LogicProperty `gorm:"foreignKey:channel,behaviour;references:channel,behaviour"json:"properties,omitempty"`
}

type BeeCommandTransferObject struct {
	Behaviour BeeNodeEndpointCluster `json:"behaviour,omitempty"`
	Command   string                 `json:"command"`
}
