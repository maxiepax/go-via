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
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ListUsers Get a list of all device classes
// @Summary Get all device classes
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Success 200 {array} models.User
// @Failure 500 {object} models.APIError
// @Router /device_classes [get]
func ListUsers(c *gin.Context) {
	var items []models.User
	if res := db.DB.Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	c.JSON(http.StatusOK, items) // 200
}

// GetUser Get an existing device class
// @Summary Get an existing device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param  id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/{id} [get]
func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.User
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

// SearchUser Search for an device class
// @Summary Search for an device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param item body models.User true "Fields to search for"
// @Success 200 {object} models.User
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/search [post]
func SearchUser(c *gin.Context) {
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
	var item models.User
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

// CreateUser Create a new device class
// @Summary Create a new device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param item body models.UserForm true "Add an device class"
// @Success 200 {object} models.User
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes [post]
func CreateUser(c *gin.Context) {
	var form models.UserForm

	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	item := models.User{UserForm: form}

	// hash and salt the plaintext password
	hp := HashAndSalt([]byte(item.Password))
	item.Password = hp
	if res := db.DB.Create(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// UpdateUser Update an existing device class
// @Summary Update an existing device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param  id path int true "User ID"
// @Param  item body models.UserForm true "Update an ip device class"
// @Success 200 {object} models.User
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/{id} [patch]
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the form data
	var form models.UserForm
	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.User
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Merge the item and the form data
	if err := mergo.Merge(&item, models.User{UserForm: form}, mergo.WithOverride); err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
	}

	// hash and salt the plaintext password
	hp := HashAndSalt([]byte(item.Password))
	item.Password = hp

	// Save it
	if res := db.DB.Save(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// DeleteUser Remove an existing device class
// @Summary Remove an existing device class
// @Tags device_classes
// @Accept  json
// @Produce  json
// @Param  id path int true "User ID"
// @Success 204
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /device_classes/{id} [delete]
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.User
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

// functions to hash and compare passwords

func HashAndSalt(pwd []byte) string {
	// Generate hashed and salted password
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("couldnt salt and hash password")
	}
	return string(hash)
}
func ComparePasswords(hashedPwd string, plainPwd []byte, username string) bool {
	// compare a password to the hashed and salted value
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"username": username,
			"error":    err,
		}).Error("invalid password supplied")
		return false
	}
	return true
}