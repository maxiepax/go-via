package models

import "time"

type Theme struct {
	ID        int        `json:"id" gorm:"primary_key"`
	ImageData []byte     `json:"image_data" gorm:"type:blob"`
	MimeType  string     `json:"mime_type" gorm:"type:varchar(64)"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
