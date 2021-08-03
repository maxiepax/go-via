package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ListGroups Get a list of all groups
// @Summary Get all groups
// @Tags groups
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Group
// @Failure 500 {object} models.APIError
// @Router /groups [get]
func ListGroups(c *gin.Context) {
	var items []models.Group
	if res := db.DB.Preload("Pool").Preload("Option").Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	items.Password = ""
	c.JSON(http.StatusOK, items) // 200
}

// GetGroup Get an existing group
// @Summary Get an existing group
// @Tags groups
// @Accept  json
// @Produce  json
// @Param  id path int true "Group ID"
// @Success 200 {object} models.Group
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /groups/{id} [get]
func GetGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Group
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

// CreateGroup Create a new groups
// @Summary Create a new group
// @Tags groups
// @Accept  json
// @Produce  json
// @Param item body models.GroupForm true "Add ip group"
// @Success 200 {object} models.Group
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /groups [post]
func CreateGroup(c *gin.Context) {
	var form models.GroupForm

	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	item := models.Group{GroupForm: form}

	//remove whitespaces surrounding comma kickstart file breaks otherwise.
	item.DNS = strings.Join(strings.Fields(item.DNS), "")
	item.NTP = strings.Join(strings.Fields(item.NTP), "")

	if res := db.DB.Create(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	// Load a new version with relations
	if res := db.DB.Preload("Pool").First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200

	logrus.WithFields(logrus.Fields{
		"Name":     item.Name,
		"DNS":      item.DNS,
		"NTP":      item.NTP,
		"Image ID": item.ImageID,
		"Pool ID":  item.PoolID,
	}).Debug("group")
}

// UpdateGroup Update an existing group
// @Summary Update an existing group
// @Tags groups
// @Accept  json
// @Produce  json
// @Param  id path int true "Group ID"
// @Param  item body models.GroupForm true "Update an group"
// @Success 200 {object} models.Group
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /groups/{id} [patch]
func UpdateGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the form data
	var form models.GroupForm
	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Group
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Merge the item and the form data
	if err := mergo.Merge(&item, models.Group{GroupForm: form}, mergo.WithOverride); err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
	}

	//remove whitespaces surrounding comma kickstart file breaks otherwise.
	item.DNS = strings.Join(strings.Fields(item.DNS), "")
	item.NTP = strings.Join(strings.Fields(item.NTP), "")

	// Save it
	if res := db.DB.Preload("Pool").Save(&item); res.Error != nil {
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

// DeleteGroup Remove an existing group
// @Summary Remove an existing group
// @Tags groups
// @Accept  json
// @Produce  json
// @Param  id path int true "Group ID"
// @Success 204
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /groups/{id} [delete]
func DeleteGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Group
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

func crypt_sha512(pass string) string {
	crypt := crypt.SHA512.New()
	ret, _ := crypt.Generate([]byte("secret"), []byte(""))
	return ret
}
