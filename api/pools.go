package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"gorm.io/gorm"
)

// ListPools Get a list of all pools
// @Summary Get all pools
// @Tags pools
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Pool
// @Failure 500 {object} models.APIError
// @Router /pools [get]
func ListPools(c *gin.Context) {
	var items []models.Pool
	if res := db.DB.Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	c.JSON(http.StatusOK, items) // 200
}

// GetPool Get an existing pool
// @Summary Get an existing pool
// @Tags pools
// @Accept  json
// @Produce  json
// @Param  id path int true "Pool ID"
// @Success 200 {object} models.Pool
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /pools/{id} [get]
func GetPool(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Pool
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// SearchPool Search for an pool
// @Summary Search for an pool
// @Tags pools
// @Accept  json
// @Produce  json
// @Param item body models.Pool true "Fields to search for"
// @Success 200 {object} models.Pool
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /pools/search [post]
func SearchPool(c *gin.Context) {
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
	var item models.Pool
	if res := query.First(&item); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// CreatePool Create a new pool
// @Summary Create a new pool
// @Tags pools
// @Accept  json
// @Produce  json
// @Param item body models.PoolForm true "Add ip pool"
// @Success 200 {object} models.Pool
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /pools [post]
func CreatePool(c *gin.Context) {
	var form models.PoolForm

	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	item := models.Pool{PoolForm: form}

	if res := db.DB.Create(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	for i, value := range item.DNS {
		var opt models.Option
		opt.PoolID = item.ID
		opt.OpCode = 6
		opt.Data = value
		opt.Priority = i + 1

		if res := db.DB.Create(&opt); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}
		spew.Dump(opt.ID)
	}

	c.JSON(http.StatusOK, item) // 200
}

// GetNextFreeIP Get the next free lease from a pool
// @Summary Get the next free lease from a pool
// @Tags pools
// @Accept  json
// @Produce  json
// @Param  id path int true "Pool ID"
// @Success 200 {object} models.Address
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /pools/{id}/next [get]
func GetNextFreeIP(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.PoolWithAddresses
	if res := db.DB.Table("pools").Preload("Addresses", "reserved OR expires > NOW()").First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	ip, err := item.Next()
	if err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
		return
	}

	resp := models.Address{
		AddressForm: models.AddressForm{
			IP: ip.String(),
		},
	}

	c.JSON(http.StatusOK, resp) // 200
}

// GetPoolByRelay Get an existing pool by the relay host IP
// @Summary Get an existing pool by the relay host IP
// @Tags pools
// @Accept  json
// @Produce  json
// @Param  relay path string true "Relay host IP"
// @Success 200 {object} models.Pool
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /relay/{relay} [get]
func GetPoolByRelay(c *gin.Context) {
	relay := c.Param("relay")
	if relay == "" {
		Error(c, http.StatusBadRequest, fmt.Errorf("relay host parameter is missing")) // 400
		return
	}

	// Load the item
	
	item, err := FindPool(relay)
	if err != nil {
		Error(c, http.StatusNotFound, fmt.Errorf("not found"))
		return
	}

	/*
	var item models.Pool
	if res := db.DB.Where("INET_ATON(net_address) = INET_ATON(?) & ((POWER(2, netmask)-1) <<(32-netmask))", relay).First(&item); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}
	*/

	c.JSON(http.StatusOK, item) // 200
}

// UpdatePool Update an existing pool
// @Summary Update an existing pool
// @Tags pools
// @Accept  json
// @Produce  json
// @Param  id path int true "Pool ID"
// @Param  item body models.PoolForm true "Update an ip pool"
// @Success 200 {object} models.Pool
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /pools/{id} [patch]
func UpdatePool(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the form data
	var form models.PoolForm
	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Pool
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Merge the item and the form data
	if err := mergo.Merge(&item, models.Pool{PoolForm: form}, mergo.WithOverride); err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
	}

	item.OnlyServeReserved = form.OnlyServeReserved

	// Save it
	if res := db.DB.Save(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// DeletePool Remove an existing pool
// @Summary Remove an existing pool
// @Tags pools
// @Accept  json
// @Produce  json
// @Param  id path int true "Pool ID"
// @Success 204
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /pools/{id} [delete]
func DeletePool(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Pool
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Save it
	if res := db.DB.Delete(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusNoContent, gin.H{}) //204
}

func FindPool(ip string) (*models.PoolWithAddresses, error) {
	var pools []models.Pool
	if res := db.DB.Table("pools").Find(&pools); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	var pool models.PoolWithAddresses
	for _, v := range pools {
		_, ipv4Net, err := net.ParseCIDR(ip + "/" + strconv.Itoa(v.Netmask))
		if err != nil {
			continue
		}
		spew.Dump(ipv4Net.String())
		spew.Dump(v.NetAddress)
		if ipv4Net.String() == v.NetAddress {
			pool.Pool = v
			break
		}
	}

	if pool.ID == 0 {
		return nil, fmt.Errorf("no matching pool found")
	}

	if res := db.DB.Table("pools").Preload("Addresses").First(&pool, pool.ID); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no matching pool found")
		}
		return nil, res.Error
	}
	return &pool, nil
}
