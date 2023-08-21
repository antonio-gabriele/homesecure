package infra

import (
	_ "embed"
	"time"
)

type LogicProperty struct {
	Channel   string    `gorm:"primaryKey"json:"channel,omitempty"`
	Behaviour string    `gorm:"primaryKey"json:"behaviour,omitempty"`
	Property  string    `gorm:"primaryKey"json:"property,omitempty"`
	UpdateAt  time.Time `json:"updatedAt,omitempty"`
	Value     string    `json:"value,omitempty"`
}

type LogicCommand struct {
	Channel   string `gorm:"primaryKey"json:"channel,omitempty"`
	Behaviour string `gorm:"primaryKey"json:"behaviour,omitempty"`
	Command   string `gorm:"primaryKey"json:"command,omitempty"`
}

type LogicLink struct {
	Channel1    string `gorm:"primaryKey"json:"channel1,omitempty"`
	Behaviour1  string `gorm:"primaryKey"json:"behaviour1,omitempty"`
	Channel2    string `gorm:"primaryKey"json:"channel2,omitempty"`
	Behaviour2  string `gorm:"primaryKey"json:"behaviour2,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	Provisioned bool   `json:"provisioned,omitempty"`
}
