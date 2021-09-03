package models

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/maxiepax/go-via/db"
	"gorm.io/gorm"
)

type PoolForm struct {
	Name             string `json:"name" gorm:"type:varchar(255);not null" binding:"required" `
	StartAddress     string `json:"start_address" gorm:"type:varchar(15);not null" binding:"required" `
	EndAddress       string `json:"end_address" gorm:"type:varchar(15);not null" binding:"required" `
	Netmask          int    `json:"netmask" gorm:"type:integer;not null" binding:"required" `
	LeaseTime        int    `json:"lease_time" gorm:"type:bigint" binding:"required" `
	Gateway          string `json:"gateway" gorm:"type:varchar(15)" binding:"required" `
	OnlyServeReimage bool   `json:"only_serve_reimage" gorm:"type:boolean"`
	//DNS               []string `json:"dns" gorm:"-" "type:varchar(255)`

	AuthorizedVlan int    `json:"authorized_vlan" gorm:"type:bigint"`
	ManagedRef     string `json:"managed_reference"`
}

type Pool struct {
	ID int `json:"id" gorm:"primary_key"`

	NetAddress string `json:"net_address" gorm:"type:varchar(15);not null"`
	PoolForm

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type PoolWithAddresses struct {
	Pool
	Addresses []Address `json:"address,omitempty" gorm:"foreignkey:PoolID"`
}

func (p *Pool) BeforeCreate(tx *gorm.DB) error {
	return p.BeforeSave(tx)
}

func (p *Pool) BeforeSave(tx *gorm.DB) error {
	if p.Netmask < 1 || p.Netmask > 32 {
		return fmt.Errorf("invalid netmask")
	}

	cidrMask := "/" + strconv.Itoa(p.Netmask)
	_, startNet, err := net.ParseCIDR(p.StartAddress + cidrMask)
	if err != nil {
		return err
	}

	_, endNet, err := net.ParseCIDR(p.EndAddress + cidrMask)
	if err != nil {
		return err
	}

	if !startNet.IP.Equal(endNet.IP) {
		return fmt.Errorf("start and end address do not belong to the same network")
	}

	p.NetAddress = startNet.IP.String()

	return nil
}

// Next returns the next free address in the pool (that is not reserved nor already leased)
func (p *PoolWithAddresses) Next() (ip net.IP, err error) {
	cidrMask := "/" + strconv.Itoa(p.Netmask)
	startIP, startNet, err := net.ParseCIDR(p.StartAddress + cidrMask)
	if err != nil {
		return nil, err
	}

	endIP, _, err := net.ParseCIDR(p.EndAddress + cidrMask)
	if err != nil {
		return nil, err
	}

	if startIP.IsUnspecified() {
		return nil, fmt.Errorf("start address is unspecified")
	}

	for ip := startIP; startNet.Contains(ip); next(ip) {
		if ip.IsMulticast() || ip.IsLoopback() {
			continue
		}

		if err := p.IsAvailable(ip); err == nil {
			return ip, nil
		}

		if ip.Equal(endIP) {
			break
		}
	}

	return nil, fmt.Errorf("could not find a free address")
}

func (p *PoolWithAddresses) IsAvailable(ip net.IP) error {
	return p.IsAvailableExcept(ip, "")
}

func (p *PoolWithAddresses) Contains(ip net.IP) (bool, error) {
	cidrMask := "/" + strconv.Itoa(p.Netmask)
	_, startNet, err := net.ParseCIDR(p.StartAddress + cidrMask)
	if err != nil {
		return false, err
	}

	return startNet.Contains(ip), nil
}

func (p *PoolWithAddresses) IsAvailableExcept(ip net.IP, exclude string) error {
	ok, err := p.Contains(ip)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("does not belong to the pool")
	}

	s := ip.String()
	if s == p.Gateway {
		return fmt.Errorf("cant use the gateway address")
	}

	// Check all loaded addresses
	for _, v := range p.Addresses {
		if v.IP == s && v.Expires.After(time.Now()) && v.Mac != exclude {
			return fmt.Errorf("already leased (%d)", v.ID)
		}
	}

	// Check reservations as well
	var reservations []Address
	db.DB.Where("ip = ? AND reimage", s).Find(&reservations)
	for _, v := range reservations {
		if v.IP == s && v.Mac != exclude {
			return fmt.Errorf("already reserved")
		}
	}

	return nil
}

// Credit to Mikio Hara and pnovotnak https://stackoverflow.com/questions/36166791/how-to-get-broadcast-address-of-ipv4-net-ipnet
func (p *Pool) LastAddr() (net.IP, error) {
	cidrMask := "/" + strconv.Itoa(p.Netmask)
	_, startNet, err := net.ParseCIDR(p.StartAddress + cidrMask)
	if err != nil {
		return net.IP{}, err
	}

	if startNet.IP.To4() == nil {
		return net.IP{}, fmt.Errorf("does not support IPv6 addresses")
	}
	ip := make(net.IP, len(startNet.IP.To4()))
	binary.BigEndian.PutUint32(ip, binary.BigEndian.Uint32(startNet.IP.To4())|^binary.BigEndian.Uint32(startNet.Mask))
	return ip, nil
}

func next(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
