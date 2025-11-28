package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/maxiepax/go-via/ilomapi"
	"github.com/spf13/cobra"
)

// Create the "interfaces" command
var interfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Fetch network interface details from the Redfish API",
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required parameters
		if hostAddress == "" || username == "" || password == "" {
			log.Fatalf("Error: hostaddress, username, and password are required parameters")
		}

		// Remove "https://" from host address and extract host and port
		hostAddress = strings.TrimPrefix(hostAddress, "https://")
		host, port, err := net.SplitHostPort(hostAddress)
		if err != nil {
			log.Fatalf("Error parsing host address: %v", err)
		}

		// Call the Redfish API
		ifaceConfigs, err := ilomapi.NewRedFishApi(host, port, username, password).GetInterfaces(!all)
		if err != nil {
			log.Fatalf("Error initializing Redfish API: %v", err)
		}

		// Print the interface configurations
		for _, iface := range ifaceConfigs {
			fmt.Printf("Interface Name: %s\n", iface.IfaceName)
			fmt.Printf("IP Address: %s\n", iface.IpAddress)
			fmt.Printf("MAC Address: %s\n", iface.MacAddress)
			fmt.Printf("Speed: %s\n", iface.Speed)
			fmt.Printf("Status: %s\n", iface.Status)
			fmt.Println()
		}
	},
}
