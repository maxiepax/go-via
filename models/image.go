package models

import (
	"time"
)

type ImageForm struct {
	ISOImage string `json:"iso_image" gorm:"type:varchar(255)"`
	Path     string `json:"path" gorm:"type:varchar(255)"`
	Hash     string `json:"hash" gorm:-"`
}

type Image struct {
	ID int `json:"id" gorm:"primary_key"`

	ImageForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
