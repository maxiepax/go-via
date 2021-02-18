package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"gorm.io/gorm"
)

// ListDeviceClasses Get a list of all device classes
// @Summary Get all device classes
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Success 200 {array} models.DeviceClass
// @Failure 500 {object} models.APIError
// @Router /device_classes [get]
func ListDeviceClasses(c *gin.Context) {
	var items []models.DeviceClass
	if res := db.DB.Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	c.JSON(http.StatusOK, items) // 200
}

// GetDeviceClass Get an existing device class
// @Summary Get an existing device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param  id path int true "DeviceClass ID"
// @Success 200 {object} models.DeviceClass
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/{id} [get]
func GetDeviceClass(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.DeviceClass
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

// SearchDeviceClass Search for an device class
// @Summary Search for an device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param item body models.DeviceClass true "Fields to search for"
// @Success 200 {object} models.DeviceClass
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/search [post]
func SearchDeviceClass(c *gin.Context) {
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
	var item models.DeviceClass
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

// CreateDeviceClass Create a new device class
// @Summary Create a new device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param item body models.DeviceClassForm true "Add an device class"
// @Success 200 {object} models.DeviceClass
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes [post]
func CreateDeviceClass(c *gin.Context) {
	var form models.DeviceClassForm

	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	item := models.DeviceClass{DeviceClassForm: form}

	if res := db.DB.Create(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// UpdateDeviceClass Update an existing device class
// @Summary Update an existing device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param  id path int true "DeviceClass ID"
// @Param  item body models.DeviceClassForm true "Update an ip device class"
// @Success 200 {object} models.DeviceClass
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/{id} [patch]
func UpdateDeviceClass(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the form data
	var form models.DeviceClassForm
	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.DeviceClass
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Merge the item and the form data
	if err := mergo.Merge(&item, models.DeviceClass{DeviceClassForm: form}, mergo.WithOverride); err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
	}

	// Save it
	if res := db.DB.Save(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// DeleteDeviceClass Remove an existing device class
// @Summary Remove an existing device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param  id path int true "DeviceClass ID"
// @Success 204
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/{id} [delete]
func DeleteDeviceClass(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.DeviceClass
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
