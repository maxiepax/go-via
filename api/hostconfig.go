package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/ilomapi"
	"github.com/sirupsen/logrus"
)

// CheckIP handles the POST request for checking if the ip is available under given port
func HostConfig(c *gin.Context) {

	// Extract query parameters
	iloIpAddr := c.Query("iloIpAddr")
	port := c.Query("port")
	apiFlavour := c.Query("apiFlavour")
	username := c.Query("username")
	password := c.Query("password")
	var api ilomapi.IlomApi

	// Log the parameters for debugging
	logrus.WithFields(logrus.Fields{
		"iloIpAddr": iloIpAddr,
		"port":      port,
	}).Info("Received /hostconfig request")

	// Validate parameters
	if iloIpAddr == "" || port == "" || apiFlavour == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "iloIpAddr, apiFlavour and port are required parameters",
		})
		return
	}

	// Get the IP and mac for every Network Interface

	switch apiFlavour {

	case "redfish":
		// Call the Redfish API to get the host configuration
		api = ilomapi.NewRedFishApi(iloIpAddr, port, username, password)
	}

	hostConfig, err := api.GetHostConfig()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"iloIpAddr": iloIpAddr,
			"port":      port,
			"error":     err.Error(),
		}).Error("Failed to get host configuration")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get host configuration",
			"message": err.Error(),
		})
		return
	}
	// Log the host configuration
	logrus.WithFields(logrus.Fields{
		"iloIpAddr":  iloIpAddr,
		"port":       port,
		"hostConfig": hostConfig,
	}).Info("Host configuration retrieved successfully")
	// Return the host configuration as JSON
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Host configuration fetched successfully",
		"hostConfig": hostConfig,
		"iloIpAddr":  iloIpAddr,
		"port":       port,
		"apiFlavour": apiFlavour,
	})

}
