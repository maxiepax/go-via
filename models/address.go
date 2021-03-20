package models

import (
	"time"
)

type AddressForm struct {
	IP         string    `json:"ip" gorm:"type:varchar(15);not null;index:uniqIp,unique"`
	Mac        string    `json:"mac" gorm:"type:varchar(17);not null"`
	Hostname   string    `json:"hostname" gorm:"type:varchar(255)"`
	Domain     string    `json:"domain" gorm:"type:varchar(255)"`
	Reserved   bool      `json:"reserved" gorm:"type:bool;index:uniqIp,unique"`
	PoolID     NullInt32 `json:"pool_id" gorm:"type:BIGINT" swaggertype:"integer"`
	ManagedRef string    `json:"managed_reference"`
	GroupID    NullInt32 `json:"group_id" gorm:"type:BIGINT" swaggertype:"integer"`
}

type Address struct {
	ID int `json:"id" gorm:"primary_key"`

	Pool  Pool  `json:"pool" gorm:"foreignkey:PoolID"`
	Group Group `json:"group" gorm:"foreignkey:PoolID"`

	AddressForm

	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`

	// DHCP parameters
	LastSeenRelay  string    `json:"last_seen_relay" gorm:"type:varchar(15)"`
	MissingOptions string    `json:"missing_options" gorm:"type:varchar(255)"`
	Expires        time.Time `json:"expires_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
