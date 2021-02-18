package models

import (
	"time"
)

type ImageForm struct {
	ISO    string `json:"iso" gorm:"type:varchar(255)"`
	Path   string `json:"path" gorm:"type:varchar(255)"`
	Active bool   `json:"active" gorm:"type:bool"`
}

type Image struct {
	ID int `json:"id" gorm:"primary_key"`

	ImageForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
