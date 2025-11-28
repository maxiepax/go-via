package models

import (
	"time"
)

type UserForm struct {
	Username string `json:"username" gorm:"type:varchar(255)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	Email    string `json:"email" gorm:"type:varchar(255)"`
	Comment  string `json:"comment" gorm:"type:varchar(255)"`
}

type User struct {
	ID int `json:"id" gorm:"primary_key"`

	UserForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UserLogin struct {
	Username string `json:"username" gorm:"type:varchar(255)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
}
