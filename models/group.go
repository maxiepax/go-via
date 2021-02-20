package models

import (
	"time"
)

type GroupForm struct {
	Name     string `json:"name" gorm:"type:varchar(255)"`
	DNS      string `json:"dns" gorm:"type:varchar(255)"`
	NTP      string `json:"ntp" gorm:"type:varchar(255)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	ImageID  string `json:"image_id" gorm:"type:BIGINT"`
}

type Group struct {
	ID int `json:"id" gorm:"primary_key"`

	GroupForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
