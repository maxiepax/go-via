package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	

	"github.com/google/gopacket/layers"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/api"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func processPacket(t layers.DHCPMsgType, req *layers.DHCPv4, sourceNet net.IP, ip net.IP) (resp *layers.DHCPv4, err error) {
	switch t {
	case layers.DHCPMsgTypeDiscover:
		return processDiscover(req, sourceNet, ip)
	case layers.DHCPMsgTypeRequest:
		return processRequest(req, sourceNet, ip)
	case layers.DHCPMsgTypeRelease:
		//return "Release"
	case layers.DHCPMsgTypeInform:
		return nil, fmt.Errorf("ignored, inform type")
	case layers.DHCPMsgTypeDecline:
		return processDecline(req, sourceNet, ip)

	case layers.DHCPMsgTypeUnspecified:
		return nil, fmt.Errorf("ignored, unspecified type")
	case layers.DHCPMsgTypeOffer:
		return nil, fmt.Errorf("ignored, offer type")
	case layers.DHCPMsgTypeAck:
		return nil, fmt.Errorf("ignored, ack type")
	case layers.DHCPMsgTypeNak:
		return nil, fmt.Errorf("ignored, nak type")
	}

	return nil, fmt.Errorf("unknown dhcp request type")
}

func processDiscover(req *layers.DHCPv4, sourceNet net.IP, ip net.IP) (resp *layers.DHCPv4, err error) {
	// Find all reserved addresses that is not yet assigned a pool
	var reservedAddresses []models.Address
	if res := db.DB.Where("pool_id IS NULL").Where("reserved = 1").Find(&reservedAddresses); res.Error != nil {
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, res.Error
		}
	}

	// Figure out and get the pool
	/*
	var pool models.PoolWithAddresses
	if res := db.DB.Table("pools").Preload("Addresses").Where("INET_ATON(net_address) = INET_ATON(?) & ((POWER(2, netmask)-1) <<(32-netmask))", sourceNet.String()).First(&pool); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	*/

	/*
	var pools []models.Pool
	if res := db.DB.Table("pools").Find(&pool); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	var pool models.PoolWithAddresses
	for _, v := range pools {
		_, ipv4Net, err := net.ParseCIDR(sourceNet.String() + "/" + v.Netmask)
		if err != nil {
			continue
		}
		if ipv4Net.String() = v.NetAddress {
			pool.Pool = v
			break
		}
	}

	if pool.ID = 0 {
		return nil, fmt.Errorf("no matching pool found")
	}

	if res := db.DB.Table("pools").Preload("Addresses").First(&pool, pool.ID); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	*/

	pool, err := api.FindPool(sourceNet.String())
	if err != nil {
		return nil, err
	}

	// Make a list of all reserved and pool addresses
	addresses := append(reservedAddresses, pool.Addresses...)

	// Search in the list for our mac address
	var leaseIP net.IP
	var lease *models.Address
	for _, v := range addresses {
		// Make sure the reserved IP is within the pool
		parsedIp := net.ParseIP(v.IP)
		ok, _ := pool.Contains(parsedIp)

		// Check so we havent given someone else this IP
		err := pool.IsAvailableExcept(parsedIp, req.ClientHWAddr.String())

		if v.Mac == req.ClientHWAddr.String() && ok && err == nil {
			leaseIP = parsedIp
			lease = &v
			break
		}
	}

	// Dont answer pools with "only serve requested" flag set
	if pool.OnlyServeReserved && (lease == nil || !lease.Reserved) {
		return nil, fmt.Errorf("ignored because mac address is missing from reserved addresses")
	}

	if leaseIP == nil {
		leaseIP, err = pool.Next()
		if err != nil {
			return nil, err
		}
	}

	resp = &layers.DHCPv4{
		Operation:    layers.DHCPOpReply,
		HardwareType: layers.LinkTypeEthernet,
		Xid:          req.Xid,
		YourClientIP: leaseIP,
		RelayAgentIP: req.RelayAgentIP,
		ClientHWAddr: req.ClientHWAddr,
	}

	resp.Options = append(resp.Options, layers.NewDHCPOption(layers.DHCPOptMessageType, []byte{byte(layers.DHCPMsgTypeOffer)}))

	AddOptions(req, resp, *pool, lease, ip)

	return resp, nil
}

func processRequest(req *layers.DHCPv4, sourceNet net.IP, ip net.IP) (*layers.DHCPv4, error) {
	/*if opt82, ok := option82.Decode(req); ok {
		spew.Dump(opt82)
	}*/

	// Find all reserved addresses that is not yet assigned a pool
	var reservedAddresses []models.Address
	if res := db.DB.Where("pool_id IS NULL").Where("reserved = 1").Find(&reservedAddresses); res.Error != nil {
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, res.Error
		}
	}

	// Figure out and get the pool
	pool, err := api.FindPool(sourceNet.String())
	if err != nil {
		return nil, err
	}

	/*
	var pool models.PoolWithAddresses
	if res := db.DB.Table("pools").Preload("Addresses").Where("INET_ATON(net_address) = INET_ATON(?) & ((POWER(2, netmask)-1) <<(32-netmask))", sourceNet.String()).First(&pool); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	*/

	// Make a list of all reserved and pool addresses
	addresses := append(reservedAddresses, pool.Addresses...)

	// Extract the requested IP
	var requestedIP net.IP = req.ClientIP
	for _, v := range req.Options {
		if v.Type == layers.DHCPOptRequestIP {
			requestedIP = net.IP(v.Data)
		}
	}

	// Start building the response
	resp := &layers.DHCPv4{
		Operation:    layers.DHCPOpReply,
		HardwareType: layers.LinkTypeEthernet,
		Xid:          req.Xid,
		RelayAgentIP: req.RelayAgentIP,
		ClientHWAddr: req.ClientHWAddr,
	}

	// Try to find the lease in our address list
	var lease *models.Address
	for _, v := range addresses {
		// Check so the IP is part of the pool
		parsedIp := net.ParseIP(v.IP)
		ok, _ := pool.Contains(parsedIp)

		// Check so we havent given someone else this IP
		err := pool.IsAvailableExcept(parsedIp, req.ClientHWAddr.String())

		if v.Mac == req.ClientHWAddr.String() && v.IP != requestedIP.String() && v.Expires.After(time.Now()) && ok && err == nil {
			logrus.WithFields(logrus.Fields{
				"pool":      pool.ID,
				"expected":  v.IP,
				"requested": requestedIP.String(),
			}).Warn("dhcp: wrong ip requested")
			resp.Options = append(resp.Options, layers.NewDHCPOption(layers.DHCPOptMessageType, []byte{byte(layers.DHCPMsgTypeNak)}))
			return resp, nil
		}

		if v.Mac == req.ClientHWAddr.String() {
			foundLease := models.Address(v)
			lease = &foundLease
		}
	}

	// Check if the requested IP is available
	if lease == nil || lease.IP != requestedIP.String() {
		if err := pool.IsAvailable(requestedIP); err != nil {
			logrus.WithFields(logrus.Fields{
				"pool":      pool.ID,
				"requested": requestedIP.String(),
				"err":       err,
			}).Warnf("dhcp: the requested ip is not available")
			resp.Options = append(resp.Options, layers.NewDHCPOption(layers.DHCPOptMessageType, []byte{byte(layers.DHCPMsgTypeNak)}))
			return resp, nil
		}
	}

	// Make sure the address isnt already used
	if lease != nil {
		if err := pool.IsAvailableExcept(requestedIP, req.ClientHWAddr.String()); err != nil {
			logrus.WithFields(logrus.Fields{
				"pool":      pool.ID,
				"requested": requestedIP.String(),
				"err":       err,
			}).Warnf("dhcp: the requested ip is not available (used by someone else)")
			resp.Options = append(resp.Options, layers.NewDHCPOption(layers.DHCPOptMessageType, []byte{byte(layers.DHCPMsgTypeNak)}))
			return resp, nil
		}
	}

	// Dont answer pools with "only serve requested" flag set
	if pool.OnlyServeReserved && (lease == nil || !lease.Reserved) {
		return nil, fmt.Errorf("ignored because mac address is missing from reserved addresses")
	}

	// Its a new lease!
	if lease == nil {
		lease = &models.Address{
			AddressForm: models.AddressForm{
				Mac:      req.ClientHWAddr.String(),
				Hostname: "-",
				Reserved: false,
			},
		}
	}

	// Respond with the same hostname
	for _, v := range req.Options {
		if v.Type == layers.DHCPOptHostname {
			lease.Hostname = string(v.Data)
		}
	}

	resp.YourClientIP = requestedIP

	resp.Options = append(resp.Options, layers.NewDHCPOption(layers.DHCPOptMessageType, []byte{byte(layers.DHCPMsgTypeAck)}))
	AddOptions(req, resp, *pool, lease, ip)

	lease.IP = requestedIP.String()
	lease.PoolID = models.NullInt32{sql.NullInt32{int32(pool.ID), true}}
	lease.LastSeenRelay = req.RelayAgentIP.String()
	if (lease.FirstSeen == time.Time{}) {
		lease.FirstSeen = time.Now()
	}
	lease.LastSeen = time.Now()
	lease.Expires = time.Now().Add(3600 * time.Second)
	lease.MissingOptions = listMissingOptions(req, resp)

	if lease.ID == 0 {
		db.DB.Create(lease)
	} else {
		// Remove the previous record if there is any
		db.DB.Exec("DELETE FROM addresses WHERE ip=? AND reserved=0 AND expires_at <= NOW()", lease.IP)
		db.DB.Save(lease)
	}

	return resp, nil
}

func listMissingOptions(req *layers.DHCPv4, resp *layers.DHCPv4) string {
	requested := map[byte]struct{}{}
	for _, v := range req.Options {
		if v.Type == layers.DHCPOptParamsRequest {
			for _, v := range v.Data {
				requested[v] = struct{}{}
			}
		}
	}

	for _, v := range resp.Options {
		if _, ok := requested[byte(v.Type)]; ok {
			delete(requested, byte(v.Type))
		}
	}

	var list []string
	for k := range requested {
		list = append(list, strconv.Itoa(int(k)))
	}

	return strings.Join(list, ",")
}

// a IP address conflict was detected, add/update the address table to block that address from being used for a while (lease time)
func processDecline(req *layers.DHCPv4, sourceNet net.IP, ip net.IP) (*layers.DHCPv4, error) {
	/*if opt82, ok := option82.Decode(req); ok {
		spew.Dump(opt82)
	}*/

	pool, err := api.FindPool(sourceNet.String())
	if err != nil {
		return nil, err
	}

	/*
	var pool models.PoolWithAddresses
	if res := db.DB.Table("pools").Preload("Addresses").Where("INET_ATON(net_address) = INET_ATON(?) & ((POWER(2, netmask)-1) <<(32-netmask))", sourceNet.String()).First(&pool); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	*/

	var requestedIP net.IP
	for _, v := range req.Options {
		if v.Type == layers.DHCPOptRequestIP {
			requestedIP = net.IP(v.Data)
		}
	}

	// Try to find the lease in our address history
	var lease *models.Address
	for _, v := range pool.Addresses {
		if v.IP == requestedIP.To4().String() {
			lease = &v
		}
	}

	// Its an unknown device
	if lease == nil {
		lease = &models.Address{
			AddressForm: models.AddressForm{
				IP:       requestedIP.String(),
				Hostname: "-",
				Reserved: false,
			},
		}
	}

	lease.Mac = ""
	lease.PoolID = models.NullInt32{sql.NullInt32{int32(pool.ID), true}}
	lease.LastSeenRelay = req.RelayAgentIP.String()
	lease.LastSeen = time.Now()
	lease.Expires = time.Now().Add(3600 * time.Second)

	if lease.ID == 0 {
		db.DB.Create(lease)
	} else {
		db.DB.Save(lease)
	}

	return nil, nil
}

// AddOptions will try to add all requested options and the manually specified ones to the response
func AddOptions(req *layers.DHCPv4, resp *layers.DHCPv4, pool models.PoolWithAddresses, lease *models.Address, ip net.IP) error {
	var options []models.Option
	var leaseID interface{}

	if lease != nil {
		leaseID = lease.ID
	}

	// Try to find the device class
	var deviceClass models.DeviceClass
	for _, v := range req.Options {
		if v.Type == 60 { // Vendor class
			db.DB.Where("? LIKE concat('%',vendor_class,'%')", string(v.Data)).First(&deviceClass)
		}
	}

	if res := db.DB.Where("((ISNULL(pool_id) AND ISNULL(pool_id) AND ISNULL(address_id)) OR pool_id = ? OR (address_id > 0 AND address_id = ?)) AND (ISNULL(device_class_id) OR device_class_id = ?)", pool.ID, leaseID, deviceClass.ID).Order("device_class_id desc").Order("address_id desc").Order("pool_id desc").Find(&options); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return res.Error
	}

	// Group options by opcode
	byOpCode := make(map[byte][]models.Option)
	for _, v := range options {
		if byOpCode[v.OpCode] == nil {
			byOpCode[v.OpCode] = make([]models.Option, 0)
		}

		// Only add the highest level options to the list
		// The level is decided on pool_id and address_id fields
		// addess+device_class specific = 5
		// pool+device_class specific = 4
		// global+device_class = 3
		// addess specific = 2
		// pool specific = 1
		// global = 0
		if len(byOpCode[v.OpCode]) == 0 || v.Level() >= byOpCode[v.OpCode][0].Level() {
			byOpCode[v.OpCode] = append(byOpCode[v.OpCode], v)
		}
	}

	// Extract the order of the requested options
	requestedOptions := map[byte]struct{}{}
	for _, v := range req.Options {
		if v.Type == layers.DHCPOptParamsRequest {
			for _, v := range v.Data {
				requestedOptions[v] = struct{}{}
			}
		}
	}

	defaultOptions := []byte{
		byte(layers.DHCPOptT1),
		byte(layers.DHCPOptT2),
		byte(layers.DHCPOptLeaseTime),
		byte(layers.DHCPOptServerID),
	}
	for _, v := range defaultOptions {
		if _, ok := requestedOptions[v]; !ok {
			requestedOptions[v] = struct{}{}
		}
	}

	// Add the requested options to the response
	var leaseTime float64 = float64(pool.LeaseTime)
	if leaseTime == 0 {
		leaseTime = 3600
	}
	for opCode := range requestedOptions {
		if options, ok := byOpCode[opCode]; ok {
			for _, v := range options {
				dhcpOpt, _, err := v.ToDHCPOption() // TODO: fix merge
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"opcode": opCode,
						"name":   layers.DHCPOpt(opCode).String(),
						"err":    err,
					}).Error("dhcp: failed to encode dhcp option")
					continue
				}

				resp.Options = append(resp.Options, dhcpOpt)
			}
			delete(byOpCode, opCode)
			continue
		}

		// Try to generate the missing option
		code := layers.DHCPOpt(opCode)
		switch code {
		case layers.DHCPOptSubnetMask:
			resp.Options = append(resp.Options, layers.NewDHCPOption(code, net.CIDRMask(pool.Netmask, 32)))
		case layers.DHCPOptClasslessStaticRoute:
			var b bytes.Buffer
			b.Write([]byte{byte(pool.Netmask)})

			// Only write the non-zero octets.
			dstLen := (pool.Netmask + 7) / 8
			b.Write(net.ParseIP(pool.Gateway).To4()[:dstLen])

			b.Write(net.ParseIP(pool.Gateway).To4())
			resp.Options = append(resp.Options, layers.NewDHCPOption(code, b.Bytes()))
		case layers.DHCPOptRouter:
			resp.Options = append(resp.Options, layers.NewDHCPOption(code, net.ParseIP(pool.Gateway).To4()))
		case layers.DHCPOptBroadcastAddr:
			b, err := pool.LastAddr()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"opcode": opCode,
					"name":   layers.DHCPOpt(opCode).String(),
					"err":    err,
				}).Warn("dhcp: could not get broadcast address")
				continue
			}

			resp.Options = append(resp.Options, layers.NewDHCPOption(code, b))
		case layers.DHCPOptT1:
			resp.Options = append(resp.Options, models.NewUint32Option(layers.DHCPOptT1, int(leaseTime*0.5))) // renewal time
		case layers.DHCPOptT2:
			resp.Options = append(resp.Options, models.NewUint32Option(layers.DHCPOptT2, int(leaseTime*0.875))) // rebind time
		case layers.DHCPOptLeaseTime:
			resp.Options = append(resp.Options, models.NewUint32Option(layers.DHCPOptLeaseTime, int(leaseTime))) // lease time
		case layers.DHCPOptServerID:
			resp.Options = append(resp.Options, layers.NewDHCPOption(code, ip))
		default:
			// Everything failed :/
			logrus.WithFields(logrus.Fields{
				"opcode": opCode,
				"name":   layers.DHCPOpt(opCode).String(),
			}).Debug("dhcp: could not find the requested option", opCode, layers.DHCPOpt(opCode).String())
		}

	}

	// Add the remaining options (that werent requested) in the end
	for opCode, options := range byOpCode {
		for _, v := range options {
			dhcpOpt, _, err := v.ToDHCPOption() // TODO: fix merge
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"opcode": opCode,
					"name":   layers.DHCPOpt(opCode).String(),
					"err":    err,
				}).Error("dhcp: failed to encode dhcp option")
				continue
			}

			resp.Options = append(resp.Options, dhcpOpt)
		}
	}

	return nil
}
