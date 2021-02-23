package models

import (
	"time"

	"github.com/qor/media/oss"
)

type ImageForm struct {
	ISOImage string `json:"iso_image" gorm:"type:varchar(255)"`
	Path     string `json:"path" gorm:"type:varchar(255)"`
	Image    oss.OSS
}

type Image struct {
	ID int `json:"id" gorm:"primary_key"`

	ImageForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
