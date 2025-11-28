package api

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/ilomapi"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/maxiepax/go-via/models"
)

func StartIloHost(c *gin.Context) {
	// Log that the method has been called
	logrus.Debug("Received start request")
	// Extract query parameters and create API client
	api, _, err := createAPIClientFromParams(c)
	if err != nil {
		return
	}

	err = api.StartServer()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to start host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start host",
			"message": err.Error(),
		})
		return
	}
	// Log the host configuration
	logrus.WithFields(logrus.Fields{
		"endpoint": api.GetEndpoint(),
	}).Info("Host has been scheduled to start successfully")
	// Return the host configuration as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Host configuration fetched successfully",

		"endpoint":   api.GetEndpoint(),
		"apiFlavour": api.GetFlavour(),
	})

}

func createAPIClientFromParams(c *gin.Context) (ilomapi.IlomApi, map[string]string, error) {
	// Extract query parameters
	var params map[string]string
	var requestBody struct {
		IloIpAddr  string `json:"iloIpAddr"`
		Port       string `json:"port"`
		ApiFlavour string `json:"apiFlavour"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		VlanID     int    `json:"vlanID"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {

		fmt.Println(c.Request.PostForm)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return nil, params, errors.New("Invalid request body: " + err.Error())
	}

	iloIpAddr := requestBody.IloIpAddr
	port := requestBody.Port
	apiFlavour := requestBody.ApiFlavour
	username := requestBody.Username
	password := requestBody.Password

	params = map[string]string{
		"iloIpAddr":  requestBody.IloIpAddr,
		"port":       requestBody.Port,
		"apiFlavour": requestBody.ApiFlavour,
		"username":   requestBody.Username,
		"password":   requestBody.Password,
		"vlanID":     fmt.Sprintf("%d", requestBody.VlanID),
	}

	var api ilomapi.IlomApi

	// Validate parameters
	if iloIpAddr == "" || port == "" || apiFlavour == "" || username == "" || password == "" {
		if len(password) > 0 {
			password = "*****" // Mask password for security
		}
		logrus.WithFields(logrus.Fields{
			"iloIpAddr":  iloIpAddr,
			"port":       port,
			"apiFlavour": apiFlavour,
			"username":   username,
			"password":   password, // Mask password for security
		}).Warn("Missing required parameters")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "iloIpAddr, apiFlavour, port, username, and password are required parameters",
		})
		return nil, params, errors.New("Missing required parameters: iloIpAddr:" + iloIpAddr +
			" port:" + port +
			" apiFlavour:" + apiFlavour +
			" username:" + username +
			" password:" + password)
	}

	// Get the IP and mac for every Network Interface

	switch apiFlavour {

	case "redfish":
		// Call the Redfish API to get the host configuration
		api = ilomapi.NewRedFishApi(iloIpAddr, port, username, password)
	}
	// return the API client
	return api, params, nil
}

func ShutdownIloHost(c *gin.Context) {

	// Log that the method has been called
	logrus.Debug("Received shutdown request")

	// Extract query parameters and create API client
	api, _, err := createAPIClientFromParams(c)
	if err != nil {
		return
	}

	err = api.StopServer()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to shutdown host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to shutdown host",
			"message": err.Error(),
		})
		return
	}
	// Log the host configuration
	logrus.WithFields(logrus.Fields{
		"endpoint": api.GetEndpoint(),
	}).Debug("Host has been shut down successfully")
	// Return the host configuration as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Host configuration fetched successfully",

		"endpoint":   api.GetEndpoint(),
		"apiFlavour": api.GetFlavour(),
	})

}

func GetServerPowerStateFromIlo(c *gin.Context) {
	// Log that the method has been called
	logrus.Debug("Received get power state request")
	// Extract query parameters and create API client
	api, _, err := createAPIClientFromParams(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to create client")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create client",
			"message": err.Error(),
		})
		return
	}

	powerState, err := api.GetServerPowerState()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to get power state")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get power state",
			"message": err.Error(),
		})
		return
	}
	// Log the host configuration
	logrus.WithFields(logrus.Fields{
		"endpoint":   api.GetEndpoint(),
		"powerState": powerState.String(),
	}).Debug("Fetched power state successfully")
	// Return the host configuration as JSON
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Fetched power state successfully",
		"powerState": powerState.String(),

		"endpoint":   api.GetEndpoint(),
		"apiFlavour": api.GetFlavour(),
	})

}

func RebootIloHost(c *gin.Context) {
	// Log that the method has been called
	logrus.Debug("Received reboot request")
	// Extract query parameters and create API client
	api, _, err := createAPIClientFromParams(c)
	if err != nil {
		return
	}

	err = api.RebootServer()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to reboot host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reboot host",
			"message": err.Error(),
		})
		return
	}

	// Log success and return response
	logrus.WithFields(logrus.Fields{
		"endpoint": api.GetEndpoint(),
	}).Info("Host has been rebooted successfully")
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Host rebooted successfully",
		"endpoint":   api.GetEndpoint(),
		"apiFlavour": api.GetFlavour(),
	})
}

func OneTimeBoot(c *gin.Context) {

	// Log that the method has been called
	logrus.Info("Received reboot request")
	// Extract query parameters and create API client
	api, _, err := createAPIClientFromParams(c)
	if err != nil {
		return
	}

	err = api.SetOneTimeHTTPBoot()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to reboot host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reboot host",
			"message": err.Error(),
		})
		return
	}

	// Log success and return response
	logrus.WithFields(logrus.Fields{
		"endpoint": api.GetEndpoint(),
	}).Info("Host has been rebooted successfully")
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Host rebooted successfully",
		"endpoint":   api.GetEndpoint(),
		"apiFlavour": api.GetFlavour(),
	})
}

func SetVLANID(c *gin.Context) {
	// Log that the method has been called
	logrus.Debug("Received set VLANID request")

	// Extract query parameters and create API client
	api, params, err := createAPIClientFromParams(c)
	if err != nil {
		return
	}

	vlanID, err := strconv.Atoi(params["vlanID"])

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to parse VLAN ID")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse VLAN ID",
			"message": err.Error(),
		})
		return
	}
	// Set VLAN ID
	err = api.SetVLANID(vlanID)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to reboot host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reboot host",
			"message": err.Error(),
		})
		return
	}

	err = api.SetOneTimeHTTPBoot()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to set one Time Boot host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set one Time Boot host",
			"message": err.Error(),
		})
		return
	}

	err = api.RebootServer()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": api.GetEndpoint(),
			"error":    err.Error(),
		}).Error("Failed to trigger reboot host")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to trigger reboot host",
			"message": err.Error(),
		})
		return
	}

	// Log success and return response
	logrus.WithFields(logrus.Fields{
		"endpoint": api.GetEndpoint(),
	}).Info("Host has been configured with VLANID " + params["vlanID"] + " successfully")
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Host has been configured with VLANID " + params["vlanID"] + " successfully",
		"endpoint":   api.GetEndpoint(),
		"apiFlavour": api.GetFlavour(),
	})
}

// ListHosts Get a list of all hosts
// @Summary Get all hosts
// @Tags hosts
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Host
// @Failure 500 {object} models.APIError
// @Router /hosts [get]
func ListHosts(c *gin.Context) {
	var items []models.Host
	if res := db.DB.Preload("Pool").Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	c.JSON(http.StatusOK, items) // 200
}

// GetHost Get an existing Host
// @Summary Get an existing Host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param  id path int true "Host ID"
// @Success 200 {object} models.Host
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /hosts/{id} [get]
func GetHost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Host
	if res := db.DB.Preload("Pool").First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// SearchHost Search for an host
// @Summary Search for an host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param item body models.Host true "Fields to search for"
// @Success 200 {object} models.Host
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /hosts/search [post]
func SearchHost(c *gin.Context) {
	form := make(map[string]interface{})

	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	query := db.DB

	for k, v := range form {
		query = query.Where(k, v)
	}

	// Load the item
	var item models.Host
	if res := query.Preload("Pool").First(&item); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// CreateHost Create a new host
// @Summary Create a new host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param item body models.HostForm true "Add a host"
// @Success 200 {object} models.Host
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /hosts [post]
func CreateHost(c *gin.Context) {
	var form models.HostForm

	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	item := models.Host{HostForm: form}

	// get the pool network info to verify if this ip should be added to the pool.
	var pool models.Pool
	db.DB.First(&pool, "id = ?", form.PoolID)

	// first check if the address is even in the network.
	ip, err := netip.ParseAddr(item.IP)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("CreateHost")
		Error(c, http.StatusInternalServerError, err) // 500
		return
	}

	network, err := netip.ParsePrefix(pool.NetAddress + "/" + strconv.Itoa(pool.Netmask))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"form": form,
		}).Error("CreateHost")
		logrus.WithFields(logrus.Fields{
			"ip":      ip,
			"network": network,
			"pool-id": pool.ID,
			"pool":    pool.Name,
			"error":   err,
		}).Warn("ip validation failed")
		Error(c, http.StatusInternalServerError, err) // 500
		return
	}

	if network.Contains(ip) {
		logrus.WithFields(logrus.Fields{
			"ip":      ip,
			"network": network,
		}).Debug("ip validation successful")
	} else {
		logrus.WithFields(logrus.Fields{
			"ip":      ip,
			"network": network,
			"pool-id": pool.ID,
			"pool":    pool.Name,
			"error":   err,
		}).Warn("ip validation failed")
		Error(c, http.StatusBadRequest, fmt.Errorf("the ip address is not in the scope of the dhcp pool associated with the group - %s/%d", pool.NetAddress, pool.Netmask)) // 400
		return
	}

	// ensure the mac address is properly formated.
	mac, _ := net.ParseMAC(item.Mac)
	item.Mac = mac.String()

	// if ip address checks pass, continue to commit.
	if item.ID != 0 { // Save if its an existing item
		if res := db.DB.Save(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}
	} else { // Create a new item
		if res := db.DB.Create(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}
	}

	// Load a new version with relations
	if res := db.DB.Preload("Pool").First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200

	logrus.WithFields(logrus.Fields{
		"Hostname": item.Hostname,
		"Domain":   item.Domain,
		"IP":       item.IP,
		"MAC":      item.Mac,
		"Pool ID":  item.PoolID,
		"Group ID": item.GroupID,
	}).Debug("host")
}

// UpdateHost Update an existing host
// @Summary Update an existing host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param  id path int true "Host ID"
// @Param  item body models.HostForm true "Update a host"
// @Success 200 {object} models.Host
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /hosts/{id} [patch]
func UpdateHost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the form data
	var form models.HostForm
	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Host
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Merge the item and the form data
	if err := mergo.Merge(&item, models.Host{HostForm: form}, mergo.WithOverride); err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
	}

	// Mergo doesn't overwrite 0 or false values, force set
	item.Reimage = form.Reimage
	item.Progress = form.Progress

	// Save it
	if res := db.DB.Save(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	// Load a new version with relations
	if res := db.DB.Preload("Pool").First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// DeleteHost Remove an existing host
// @Summary Remove an existing host
// @Tags hosts
// @Accept  json
// @Produce  json
// @Param  id path int true "Host ID"
// @Success 204
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /hosts/{id} [delete]
func DeleteHost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Host
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// delete it
	if res := db.DB.Delete(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusNoContent, gin.H{}) //204
}
