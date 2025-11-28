package main

import (
	"log"
	"net"
	"strings"

	"github.com/maxiepax/go-via/ilomapi"
	"github.com/spf13/cobra"
)

var rebootServerCmd = &cobra.Command{
	Use:   "rebootServer",
	Short: "Reboot the server",
	Run: func(cmd *cobra.Command, args []string) {
		// Remove "https://" from host address and extract host and port
		hostAddress = strings.TrimPrefix(hostAddress, "https://")
		host, port, err := net.SplitHostPort(hostAddress)
		if err != nil {
			log.Fatalf("Error parsing host address: %v", err)
		}
		api := ilomapi.NewRedFishApi(host, port, username, password)
		err = api.RebootServer()
		if err != nil {
			log.Fatalf("Failed to reboot server: %v", err)
		}
		log.Println("Successfully rebooted the server")
	},
}
