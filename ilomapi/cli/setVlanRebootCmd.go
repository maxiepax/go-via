package main

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/maxiepax/go-via/ilomapi"
	"github.com/spf13/cobra"
)

var setVlanRebootCmd = &cobra.Command{
	Use:   "setVlanReboot [vlanID]",
	Short: "Set VLAN ID, set one-time HTTP boot, and reboot the server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vlanID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid VLAN ID: %v", err)
		}
		// Remove "https://" from host address and extract host and port
		hostAddress = strings.TrimPrefix(hostAddress, "https://")
		host, port, err := net.SplitHostPort(hostAddress)
		if err != nil {
			log.Fatalf("Error parsing host address: %v", err)
		}
		api := ilomapi.NewRedFishApi(host, port, username, password)
		err = api.SetVLANID(vlanID)
		if err != nil {
			log.Fatalf("Failed to set VLAN ID: %v", err)
		}
		log.Printf("Successfully set VLAN ID to %d", vlanID)

		err = api.SetOneTimeHTTPBoot()
		if err != nil {
			log.Fatalf("Failed to set one-time HTTP boot: %v", err)
		}
		log.Println("Successfully set one-time HTTP boot")

		err = api.RebootServer()
		if err != nil {
			log.Fatalf("Failed to reboot server: %v", err)
		}
		log.Println("Successfully rebooted the server")
	},
}
