package main

import (
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var hostAddress, username, password string
var all bool = false
var debug bool = false // Global debug flag

func main() {
	// Create the root command
	var rootCmd = &cobra.Command{
		Use:   "redfish-cli",
		Short: "A CLI tool for managing Redfish-enabled devices",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set log level to debug if --debug or -v is passed
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
				logrus.Debug("Debug logging enabled")
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}
		},
	}

	// Add the global --debug flag
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")

	// Add the "interfaces" command
	interfacesCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	interfacesCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	interfacesCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	interfacesCmd.Flags().BoolVarP(&all, "all", "A", false, "Show all interfaces (default: only active)")
	err := interfacesCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = interfacesCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = interfacesCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(interfacesCmd)

	// Add the "setvlanid" command
	setVlanIDCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	setVlanIDCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	setVlanIDCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	err = setVlanIDCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = setVlanIDCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = setVlanIDCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(setVlanIDCmd)

	// Add the "setOneTimeHTTPBoot" command
	setOneTimeHTTPBootCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	setOneTimeHTTPBootCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	setOneTimeHTTPBootCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	err = setOneTimeHTTPBootCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = setOneTimeHTTPBootCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = setOneTimeHTTPBootCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(setOneTimeHTTPBootCmd)

	// Add the "RebootServer" command
	rebootServerCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	rebootServerCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	rebootServerCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	err = rebootServerCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = rebootServerCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = rebootServerCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(rebootServerCmd)

	// Add the "StopServer" command
	stopServerCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	stopServerCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	stopServerCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	err = stopServerCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = stopServerCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = stopServerCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(stopServerCmd)

	// Add the "StartServer" command
	startServerCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	startServerCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	startServerCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	err = startServerCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = startServerCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = startServerCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(startServerCmd)

	// Add the "setVlanReboot" command
	setVlanRebootCmd.Flags().StringVarP(&hostAddress, "hostaddress", "H", "", "Host address (e.g., https://19.89.89.8:443)")
	setVlanRebootCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	setVlanRebootCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication")
	err = setVlanRebootCmd.MarkFlagRequired("hostaddress")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = setVlanRebootCmd.MarkFlagRequired("username")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	err = setVlanRebootCmd.MarkFlagRequired("password")
	if err != nil {
		logrus.Errorf("Error marking flag required: %v", err)
	}
	rootCmd.AddCommand(setVlanRebootCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
