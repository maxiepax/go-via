package models

import (
	"time"

	"gorm.io/datatypes"
)

type GroupForm struct {
	PoolID   int            `json:"pool_id" gorm:"type:BIGINT"`
	Name     string         `json:"name" gorm:"type:varchar(255)"`
	DNS      string         `json:"dns" gorm:"type:varchar(255)"`
	NTP      string         `json:"ntp" gorm:"type:varchar(255)"`
	Password string         `json:"password" gorm:"type:varchar(255)"`
	ImageID  int            `json:"image_id" gorm:"type:INT"`
	Ks       string         `json:"ks" gorm:"type:text"`
	Syslog   string         `json:"syslog" gorm:"type:varchar(255)"`
	Options  datatypes.JSON `json:"options" sql:"type:JSONB" swaggertype:"object,string"`
}

type Group struct {
	ID int `json:"id" gorm:"primary_key"`

	GroupForm

	Pool    *Pool     `json:"pool,omitempty" gorm:"foreignkey:PoolID"`
	Option  []Option  `json:"option,omitempty" gorm:"foreignkey:PoolID"`
	Address []Address `json:"option,omitempty" gorm:"foreignkey:GroupID"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type GroupOptions struct {
	Domain               bool   `json:"domain"`
	NTP                  bool   `json:"ntp"`
	SSH                  bool   `json:"ssh"`
	SuppressShellWarning bool   `json:"suppressshellwarning"`
	EraseDisks           bool   `json:"erasedisks"`
	BootDisk             string `json:"bootdisk" gorm:"type:varchar(255)"`
	AllowLegacyCPU       bool   `json:"allowlegacycpu"`
	Syslog               bool   `json:"syslog"`
	Certificate          bool   `json:"certificate"`
}
