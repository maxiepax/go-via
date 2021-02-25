package models

import (
	"time"
)

type GroupForm struct {
	PoolID   int    `json:"pool_id" gorm:"type:BIGINT"`
	Name     string `json:"name" gorm:"type:varchar(255)"`
	DNS      string `json:"dns" gorm:"type:varchar(255)"`
	NTP      string `json:"ntp" gorm:"type:varchar(255)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	ImageID  int    `json:"image_id" gorm:"type:INT"`
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
