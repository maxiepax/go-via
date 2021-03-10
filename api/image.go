package api

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"gorm.io/gorm"
)

// ListImages Get a list of all images
// @Summary Get all images
// @Tags images
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Image
// @Failure 500 {object} models.APIError
// @Router /images [get]
func ListImages(c *gin.Context) {
	var items []models.Image
	if res := db.DB.Find(&items); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}
	c.JSON(http.StatusOK, items) // 200
}

// GetImage Get an existing image
// @Summary Get an existing image
// @Tags images
// @Accept  json
// @Produce  json
// @Param  id path int true "Image ID"
// @Success 200 {object} models.Image
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /images/{id} [get]
func GetImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Image
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

// CreateImage Create a new images
// @Summary Create a new image
// @Tags images
// @Accept  json
// @Produce  json
// @Param item body models.ImageForm true "Add image"
// @Success 200 {object} models.Image
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /images [post]
func CreateImage(c *gin.Context) {

	f, err := c.MultipartForm()
	if err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
		return
	}

	files := f.File["file[]"]
	spew.Dump(files)

	for _, file := range files {
		//filename := "test.iso"

		filename := file.Filename
		// updatera databas med information om iso namn och path
		item := models.Image{}
		item.ISOImage = filepath.Base(file.Filename)
		item.Path = path.Join(".", "tftp", filename) //kommer ifrån main som argument
		//*

		os.MkdirAll(filepath.Dir(item.Path), os.ModePerm)

		_, err = SaveUploadedFile(file, item.Path)
		if err != nil {
			Error(c, http.StatusInternalServerError, err) // 500
			return
		}

		/*
			mime, err := mimetype.DetectFile(item.StoragePath)
			if err != nil {
				Error(c, http.StatusInternalServerError, err) // 500
				return
			}
			item.Type = mime.String()
			item.Extension = mime.Extension()
		*/

		// commita till databas
		if result := db.DB.Table("images").Create(&item); result.Error != nil {
			Error(c, http.StatusInternalServerError, result.Error) // 500
			return
		}
		c.JSON(http.StatusOK, item) // 200
	}

	//spew.Dump(item)
	//c.JSON(http.StatusOK, item) // 200
}

func SaveUploadedFile(file *multipart.FileHeader, dst string) (int64, error) {
	src, err := file.Open()
	if err != nil {
		return -1, err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return -1, err
	}
	defer out.Close()

	n, err := io.Copy(out, src)
	return n, err
}

// UpdateImage Update an existing image
// @Summary Update an existing image
// @Tags images
// @Accept  json
// @Produce  json
// @Param  id path int true "Image ID"
// @Param  item body models.ImageForm true "Update an image"
// @Success 200 {object} models.Image
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /images/{id} [patch]
func UpdateImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the form data
	var form models.ImageForm
	if err := c.ShouldBind(&form); err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Image
	if res := db.DB.First(&item, id); res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, fmt.Errorf("not found")) // 404
		} else {
			Error(c, http.StatusInternalServerError, res.Error) // 500
		}
		return
	}

	// Merge the item and the form data
	if err := mergo.Merge(&item, models.Image{ImageForm: form}, mergo.WithOverride); err != nil {
		Error(c, http.StatusInternalServerError, err) // 500
	}

	// Save it
	if res := db.DB.Save(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	// Load a new version with relations
	if res := db.DB.First(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusOK, item) // 200
}

// DeleteImage Remove an existing image
// @Summary Remove an existing image
// @Tags images
// @Accept  json
// @Produce  json
// @Param  id path int true "Image ID"
// @Success 204
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /images/{id} [delete]
func DeleteImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, err) // 400
		return
	}

	// Load the item
	var item models.Image
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
