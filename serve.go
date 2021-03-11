package main

import (
	"fmt"
	"net"

	//"github.com/davecgh/go-spew/spew"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mdlayher/raw"
	"github.com/sirupsen/logrus"
)

func serve(intf string) {
	// Select interface to used
	ifi, err := net.InterfaceByName(intf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"if":  intf,
			"err": err,
		}).Fatalf("dhcp: failed to open interface")
	}

	// Find the ip-address
	ip, ipNet, err := findIPv4Addr(ifi)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"if":  intf,
			"err": err,
		}).Fatalf("dhcp: failed to get interface IPv4 address")
	}

	mac := ifi.HardwareAddr

	// Open a raw socket using ethertype 0x0800 (IPv4)
	c, err := raw.ListenPacket(ifi, 0x0800, &raw.Config{})
	if err != nil {
		logrus.Fatalf("dhcp: failed to listen: %v", err)
	}
	defer c.Close()

	logrus.WithFields(logrus.Fields{
		"mac": mac,
		"ip":  ip,
		"int": intf,
	}).Infof("Starting dhcp server")

	// Accept frames up to interface's MTU in size
	b := make([]byte, ifi.MTU)

	// Keep reading frames
	for {
		n, src, err := c.ReadFrom(b)
		if err != nil {
			logrus.Fatalf("dhcp: failed to receive message: %v", err)
		}

		packet := gopacket.NewPacket(b[:n], layers.LayerTypeEthernet, gopacket.Default)

		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4)

		if ethLayer != nil && ipv4Layer != nil && udpLayer != nil && dhcpLayer != nil {
			eth, _ := ethLayer.(*layers.Ethernet)
			ipv4, _ := ipv4Layer.(*layers.IPv4)
			udp, _ := udpLayer.(*layers.UDP)
			req, _ := dhcpLayer.(*layers.DHCPv4)

			//spew.Dump(req)

			t := findMsgType(req)
			sourceNet := ip
			source := "broadcast"
			if ipNet != nil && !ipNet.Contains(ipv4.SrcIP) && !ipv4.SrcIP.Equal(net.IPv4zero) {
				sourceNet = ipv4.SrcIP
				source = "unicast"
			}

			if (req.RelayAgentIP != nil && !req.RelayAgentIP.Equal(net.IP{0, 0, 0, 0})) {
				sourceNet = req.RelayAgentIP
				source = "relayed"
			}

			resp, err := processPacket(t, req, sourceNet, ip)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"type":       t.String(),
					"client-mac": req.ClientHWAddr.String(),
					"source":     sourceNet.String(),
					"relay":      req.RelayAgentIP,
					"error":      err,
				}).Warnf("dhcp: failed to process %s %s", source, t)
				continue
			}

			// Copy some information from the request like option 82 (agent info) to the response
			resp.Flags = req.Flags
			for _, v := range req.Options {
				if v.Type == layers.DHCPOptClientID {
					resp.Options = append(resp.Options, v)
				}
				if v.Type == layers.DHCPOptHostname {
					resp.Options = append(resp.Options, v)
				}
				if v.Type == 82 {
					resp.Options = append(resp.Options, v)
				}
			}

			layers := buildHeaders(mac, ip, eth, ipv4, udp)
			layers = append(layers, resp)

			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}
			err = gopacket.SerializeLayers(buf, opts, layers...)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"response":   findMsgType(resp).String(),
					"client-mac": req.ClientHWAddr.String(),
					"ip":         resp.YourClientIP,
					"relay":      req.RelayAgentIP,
				}).Warnf("dhcp: failed to serialise response to %s %s", source, t)
				continue
			}

			c.WriteTo(buf.Bytes(), src)

			//spew.Dump(resp)
			logrus.WithFields(logrus.Fields{
				"response":   findMsgType(resp).String(),
				"client-mac": req.ClientHWAddr.String(),
				"ip":         resp.YourClientIP,
				"relay":      req.RelayAgentIP,
			}).Infof("dhcp: answered %s %s with %s", source, t, findMsgType(resp))
			for _, v := range resp.Options {
				logrus.Debug(v)
			}
		}
	}
}

func findMsgType(p *layers.DHCPv4) layers.DHCPMsgType {
	var msgType layers.DHCPMsgType
	for _, o := range p.Options {
		if o.Type == layers.DHCPOptMessageType {
			msgType = layers.DHCPMsgType(o.Data[0])
		}
	}

	return msgType
}

func buildHeaders(mac net.HardwareAddr, ip net.IP, srcEth *layers.Ethernet, srcIP4 *layers.IPv4, srcUDP *layers.UDP) []gopacket.SerializableLayer {
	eth := &layers.Ethernet{
		SrcMAC:       mac,
		DstMAC:       srcEth.SrcMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := &layers.IPv4{
		SrcIP:    ip,
		DstIP:    srcIP4.SrcIP,
		Version:  4,
		TOS:      0x10,
		TTL:      128,
		Protocol: layers.IPProtocolUDP,
		Flags:    layers.IPv4DontFragment,
	}

	udp := &layers.UDP{
		SrcPort: 67, // bootps
		DstPort: 67, // bootps
	}

	// Answer to broadcast address if source address is 0.0.0.0
	if srcIP4.SrcIP.Equal(net.IPv4zero) {
		ip4.DstIP = net.IPv4(255, 255, 255, 255)
		udp.DstPort = 68
	}

	udp.SetNetworkLayerForChecksum(ip4)

	return []gopacket.SerializableLayer{eth, ip4, udp}
}

func findIPv4Addr(ifi *net.Interface) (net.IP, *net.IPNet, error) {
	addrs, err := ifi.Addrs()
	if err != nil {
		return nil, nil, err
	}
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPAddr:
			if addr := v.IP.To4(); addr != nil {
				return addr, nil, nil
			}
		case *net.IPNet:
			if addr := v.IP.To4(); addr != nil {
				return addr, v, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("could not find IPv4 address")
}
