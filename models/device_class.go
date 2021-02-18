package models

import (
	"time"
)

type DeviceClassForm struct {
	Name        string `json:"name" gorm:"type:varchar(255)"`
	VendorClass string `json:"vendor_class" gorm:"type:varchar(255)"`
}

type DeviceClass struct {
	ID int `json:"id" gorm:"primary_key"`

	DeviceClassForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
