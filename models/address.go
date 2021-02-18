package models

import (
	"time"
)

type AddressForm struct {
	IP             string    `json:"ip_address" gorm:"type:varchar(15);not null;index:uniqIp,unique"`
	Mac            string    `json:"mac" gorm:"type:varchar(17);not null"`
	Hostname       string    `json:"hostname" gorm:"type:varchar(255)"`
	Reserved       bool      `json:"reserved" gorm:"type:bool;index:uniqIp,unique"`
	PoolID         NullInt32 `json:"pool_id" gorm:"type:BIGINT" swaggertype:"integer"`
	AuthorizedVlan NullInt32 `json:"authorized_vlan" gorm:"type:SMALLINT" swaggertype:"integer"`
	ManagedRef     string    `json:"managed_reference"`
	// Host parameters
	Fqdn     string `json:"fqdn" gorm:"type:varchar(255)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	ImageID  int    `json:"image_id gorm:"type:int(8)"`
}

type Address struct {
	ID int `json:"id" gorm:"primary_key"`

	Pool Pool `json:"pool" gorm:"foreignkey:PoolID"`

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
