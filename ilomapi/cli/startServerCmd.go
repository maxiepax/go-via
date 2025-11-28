package main

import (
	"log"
	"net"
	"strings"

	"github.com/maxiepax/go-via/ilomapi"
	"github.com/spf13/cobra"
)

var startServerCmd = &cobra.Command{
	Use:   "startServer",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		// Remove "https://" from host address and extract host and port
		hostAddress = strings.TrimPrefix(hostAddress, "https://")
		host, port, err := net.SplitHostPort(hostAddress)
		if err != nil {
			log.Fatalf("Error parsing host address: %v", err)
		}
		api := ilomapi.NewRedFishApi(host, port, username, password)
		err = api.StartServer()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		log.Println("Successfully started the server")
	},
}
