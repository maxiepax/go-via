package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	/*_ "github.com/GehirnInc/crypt/sha512_crypt"*/

	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/maxiepax/go-via/secrets"
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
	var items []models.NoPWGroup
	if res := db.DB.Preload("Pool").Preload("Option").Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

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
	var item models.NoPWGroup
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
func CreateGroup(key string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var form models.GroupForm

		if err := c.ShouldBind(&form); err != nil {
			Error(c, http.StatusBadRequest, err) // 400
			return
		}

		item := models.Group{GroupForm: form}

		//remove whitespaces surrounding comma kickstart file breaks otherwise
		item.DNS = strings.Join(strings.Fields(item.DNS), "")
		item.NTP = strings.Join(strings.Fields(item.NTP), "")
		item.Syslog = strings.Join(strings.Fields(item.Syslog), "")

		//validate that password fullfills the password complexity requirements
		if err := verifyPassword(form.Password); err != nil {
			Error(c, http.StatusBadRequest, err) // 400
			return
		}
		item.Password = secrets.Encrypt(item.Password, key)

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
func UpdateGroup(key string) func(c *gin.Context) {
	return func(c *gin.Context) {
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
		item.Syslog = strings.Join(strings.Fields(item.Syslog), "")

		// to avoid re-hashing the password when no new password has been supplied, check if it was supplied
		//validate that password fullfills the password complexity requirements
		if form.Password != "" {
			if err := verifyPassword(form.Password); err != nil {
				Error(c, http.StatusBadRequest, err) // 400
				return
			}

			item.Password = secrets.Encrypt(item.Password, key)
		}

		//mergo wont overwrite values with empty space. To enable removal of ntp, dns, syslog, vlan, always overwrite.
		item.Vlan = form.Vlan
		item.DNS = form.DNS
		item.NTP = form.NTP
		item.Syslog = form.Syslog
		item.BootDisk = form.BootDisk

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
	if res := db.DB.Preload("Host").First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// check if the group is empty, if it's not, deny the delete.
	if len(item.Host) < 1 {
		// Delete it
		if res := db.DB.Delete(&item); res.Error != nil {
			Error(c, http.StatusInternalServerError, res.Error) // 500
			return
		}
		c.JSON(http.StatusNoContent, gin.H{}) //204
	} else {
		c.JSON(http.StatusConflict, "the group is not empty, please delete all hosts first.")
	}

}

func verifyPassword(s string) error {
	number := false
	upper := false
	special := false
	lower := false
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			lower = true
		}
	}
	var b2i = map[bool]int8{false: 0, true: 1}
	classes := b2i[number] + b2i[upper] + b2i[special] + b2i[lower]

	if classes < 3 {
		return fmt.Errorf("you need to use at least 3 character classes (lowercase, uppercase, special and numbers)")
	}

	if len(s) < 7 {
		return fmt.Errorf("too short, should be at least 7 characters")
	}

	return nil
}
