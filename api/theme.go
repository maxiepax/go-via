package api

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"gorm.io/gorm"
)

// UploadThemeImage handles uploading a new background image for the theme
func UploadThemeImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("background")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file"})
		}
	}()

	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB"})
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType != "image/png" && mimeType != "image/jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PNG and JPG allowed"})
		return
	}

	imgData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	var theme models.Theme
	err = db.DB.First(&theme).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		theme = models.Theme{
			ImageData: imgData,
			MimeType:  mimeType,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.DB.Create(&theme)
	} else if err == nil {
		theme.ImageData = imgData
		theme.MimeType = mimeType
		theme.UpdatedAt = time.Now()
		db.DB.Save(&theme)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme image updated"})
}

// GetThemeImage serves the current background image
func GetThemeImage(c *gin.Context) {
	var theme models.Theme
	err := db.DB.First(&theme).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "No theme image set"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	c.Data(http.StatusOK, theme.MimeType, theme.ImageData)
}
