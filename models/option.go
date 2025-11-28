package models

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/google/gopacket/layers"
)

type OptionForm struct {
	PoolID        int    `json:"pool_id" gorm:"type:BIGINT"`
	HostID        int    `json:"host_id" gorm:"type:BIGINT"`
	DeviceClassID int    `json:"device_class_id" gorm:"type:BIGINT"`
	OpCode        byte   `json:"opcode" gorm:"type:SMALLINT;unsigned;not null" binding:"required" `
	Data          string `json:"data" gorm:"type:varchar(255);not null" binding:"required" `
	Priority      int    `json:"priority" gorm:"type:SMALLINT;not null" binding:"required" `
}

type Option struct {
	ID int `json:"id" gorm:"primary_key"`

	OptionForm

	Pool *Pool `json:"pool,omitempty" gorm:"foreignkey:PoolID"`
	Host *Host `json:"host,omitempty" gorm:"foreignkey:HostID"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (o Option) Level() int {
	if o.DeviceClassID > 0 {
		if o.HostID > 0 {
			return 5
		}

		if o.PoolID > 0 {
			return 4
		}

		return 3
	}

	if o.HostID > 0 {
		return 2
	}

	if o.PoolID > 0 {
		return 1
	}

	return 0
}

func (o Option) ToDHCPOption() (opt layers.DHCPOption, merge bool, err error) {
	code := layers.DHCPOpt(o.OpCode)
	switch code {
	case // string
		layers.DHCPOptHostname,
		layers.DHCPOptMeritDumpFile,
		layers.DHCPOptDomainName,
		layers.DHCPOptRootPath,
		layers.DHCPOptExtensionsPath,
		layers.DHCPOptNISDomain,
		layers.DHCPOptNetBIOSTCPScope,
		layers.DHCPOptXFontServer,
		layers.DHCPOptXDisplayManager,
		layers.DHCPOptMessage,
		layers.DHCPOptDomainSearch,
		layers.DHCPOptSIPServers,
		66, // TFTP server name
		67: // TFTP file name

		return NewStringOption(code, o.Data), false, nil
	case // net.IP
		layers.DHCPOptSubnetMask,
		layers.DHCPOptBroadcastAddr,
		layers.DHCPOptSolicitAddr:

		return NewIPOption(code, net.ParseIP(o.Data)), false, nil
	case // n*net.IP
		layers.DHCPOptRouter,
		layers.DHCPOptTimeServer,
		layers.DHCPOptNameServer,
		layers.DHCPOptDNS,
		layers.DHCPOptLogServer,
		layers.DHCPOptCookieServer,
		layers.DHCPOptLPRServer,
		layers.DHCPOptImpressServer,
		layers.DHCPOptResLocServer,
		layers.DHCPOptSwapServer,
		layers.DHCPOptNISServers,
		layers.DHCPOptNTPServers,
		layers.DHCPOptNetBIOSTCPNS,
		layers.DHCPOptNetBIOSTCPDDS:

		return NewIPOption(code, net.ParseIP(o.Data).To4()), true, nil
	case // uint16
		layers.DHCPOptBootfileSize,
		layers.DHCPOptDatagramMTU,
		layers.DHCPOptInterfaceMTU,
		layers.DHCPOptMaxMessageSize:

		i, err := strconv.Atoi(o.Data)
		if err != nil {
			return opt, false, err
		}

		return NewUint16Option(code, i), false, nil
	case // n*uint16
		layers.DHCPOptPathPlateuTableOption:

		i, err := strconv.Atoi(o.Data)
		if err != nil {
			return opt, false, err
		}

		return NewUint16Option(code, i), true, nil
	case // int32 (signed seconds from UTC)
		layers.DHCPOptTimeOffset:

		i, err := strconv.Atoi(o.Data)
		if err != nil {
			return opt, false, err
		}

		return NewInt32Option(code, i), false, nil
	case // uint32
		layers.DHCPOptT1,
		layers.DHCPOptT2,
		layers.DHCPOptLeaseTime,
		layers.DHCPOptPathMTUAgingTimeout,
		layers.DHCPOptARPTimeout,
		layers.DHCPOptTCPKeepAliveInt:

		i, err := strconv.Atoi(o.Data)
		if err != nil {
			return opt, false, err
		}

		return NewUint32Option(code, i), false, nil
	}

	return opt, false, fmt.Errorf("unsupported dhcp option type %d", o.OpCode)
}

func NewUint16Option(t layers.DHCPOpt, v int) layers.DHCPOption {
	vi := uint16(v)
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, vi)

	return layers.NewDHCPOption(t, buf)
}

func NewInt32Option(t layers.DHCPOpt, v int) layers.DHCPOption {
	vi := int32(v)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(vi))

	return layers.NewDHCPOption(t, buf)
}

func NewUint32Option(t layers.DHCPOpt, v int) layers.DHCPOption {
	vi := uint32(v)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, vi)

	return layers.NewDHCPOption(t, buf)
}

func NewStringOption(t layers.DHCPOpt, v string) layers.DHCPOption {
	return layers.NewDHCPOption(t, []byte(v))
}

func NewIPOption(t layers.DHCPOpt, v net.IP) layers.DHCPOption {
	return layers.NewDHCPOption(t, v)
}
