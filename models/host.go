package models

import (
	"time"
)

type HostForm struct {
	IP            string    `json:"ip" gorm:"type:varchar(15);not null;index:uniqIp,unique"`
	IloIP         string    `json:"ilo_ip" gorm:"type:varchar(15);index:uniqIp,unique"`
	IloUser       string    `json:"ilo_user" gorm:"type:varchar(255)"`
	IloPassword   string    `json:"ilo_password" gorm:"type:varchar(255)"`
	IloPort       string    `json:"ilo_port" gorm:"type:varchar(255)"`
	IloApiFlavour string    `json:"ilo_api_flavour" gorm:"type:varchar(255)"`
	IloFqdn       string    `json:"ilo_fqdn" gorm:"type:varchar(255)"`
	HostFqdn      string    `json:"host_fqdn" gorm:"type:varchar(255)"`
	Mac           string    `json:"mac" gorm:"type:varchar(17);not null"`
	Hostname      string    `json:"hostname" gorm:"type:varchar(255)"`
	Domain        string    `json:"domain" gorm:"type:varchar(255)"`
	Reimage       bool      `json:"reimage" gorm:"type:bool;index:uniqIp,unique"`
	PoolID        NullInt32 `json:"pool_id" gorm:"type:BIGINT" swaggertype:"integer"`
	GroupID       NullInt32 `json:"group_id" gorm:"type:BIGINT" swaggertype:"integer"`
	Progress      int       `json:"progress" gorm:"type:INT"`
	Progresstext  string    `json:"progresstext" gorm:"type:varchar(255)"`
	Ks            string    `json:"ks" gorm:"type:text"`
}

type Host struct {
	ID int `json:"id" gorm:"primary_key"`

	Pool  Pool  `json:"pool" gorm:"foreignkey:PoolID"`
	Group Group `json:"group" gorm:"foreignkey:GroupID"`

	HostForm

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
