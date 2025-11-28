package main

import (
	"log"
	"net"
	"strings"

	"github.com/maxiepax/go-via/ilomapi"
	"github.com/spf13/cobra"
)

var setOneTimeHTTPBootCmd = &cobra.Command{
	Use:   "setOneTimeHTTPBoot",
	Short: "Set one-time boot to HTTP",
	Run: func(cmd *cobra.Command, args []string) {

		// Remove "https://" from host address and extract host and port
		hostAddress = strings.TrimPrefix(hostAddress, "https://")
		host, port, err := net.SplitHostPort(hostAddress)
		if err != nil {
			log.Fatalf("Error parsing host address: %v", err)
		}
		api := ilomapi.NewRedFishApi(host, port, username, password)
		err = api.SetOneTimeHTTPBoot()
		if err != nil {
			log.Fatalf("Failed to set one-time HTTP boot: %v", err)
		}
		log.Println("Successfully set one-time HTTP boot")
	},
}
