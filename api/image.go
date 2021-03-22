package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imdario/mergo"
	"github.com/kdomanski/iso9660/util"
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

	for _, file := range files {

		filename := file.Filename

		item := models.Image{}
		item.ISOImage = filepath.Base(file.Filename)
		item.Path = path.Join(".", "tftp", filename)

		os.MkdirAll(filepath.Dir(item.Path), os.ModePerm)

		_, err = SaveUploadedFile(file, item.Path)
		if err != nil {
			Error(c, http.StatusInternalServerError, err) // 500
			return
		}

		f, err := os.Open(item.Path)
		if err != nil {
			log.Fatalf("failed to open file: %s", err)
		}
		defer f.Close()

		//strip the filextension, eg. vmware.iso = vmware
		fn := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
		//merge into filepath
		fp := path.Join(".", "tftp", fn)

		if err = util.ExtractImageToDirectory(f, fp); err != nil {
			log.Fatalf("failed to extract image: %s", err)
		}

		//remove the file
		err = os.Remove(item.Path)
		if err != nil {
			log.Fatal(err)
		}

		//update item.Path
		item.Path = fp

		if _, err := os.Stat(item.Path + "/EFI/BOOT/BOOTX64.EFI"); err == nil {
			fmt.Printf("File exists\n")
			// Open original file
			original, err := os.Open(item.Path + "/EFI/BOOT/BOOTX64.EFI")
			if err != nil {
				log.Fatal(err)
			}
			defer original.Close()

			// Create new file
			new, err := os.Create(item.Path + "/MBOOT.EFI")
			if err != nil {
				log.Fatal(err)
			}
			defer new.Close()

			// Copy source to destination
			_, err = io.Copy(new, original)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("File does not exist\n")
			Error(c, http.StatusInternalServerError, err) // 500
		}

		//update the prefix=

		// read file into []byte
		bc, err := ioutil.ReadFile(item.Path + "/BOOT.CFG")
		if err != nil {
			log.Fatal(err)
		}
		// convert []byte into string
		sc := string(bc)

		// regexp to be matched to
		rx := "prefix="

		// replace prefix with prefix=foldername
		re := regexp.MustCompile(rx)
		s := re.ReplaceAllLiteralString(sc, "prefix="+fn)
		fmt.Println(s)

		// strip the leading / from all the modules
		rx = "/"
		re = regexp.MustCompile(rx)
		s = re.ReplaceAllLiteralString(sc, "")
		fmt.Println(s)

		// save string back to file
		err = WriteToFile(item.Path+"/BOOT.CFG", s)
		if err != nil {
			log.Fatal(err)
		}

		//update the kernelopt=

		// read file into []byte
		bc, err = ioutil.ReadFile(item.Path + "/BOOT.CFG")
		if err != nil {
			log.Fatal(err)
		}
		// convert []byte into string
		sc = string(bc)

		// regexp to be matched to
		rx = "kernelopt=.*"

		// replace prefix with prefix=foldername
		re = regexp.MustCompile(rx)
		o := re.FindString(sc)
		fmt.Printf("found string %s", o)
		s = re.ReplaceAllLiteralString(sc, o+" ks://")
		fmt.Println(s)

		// save string back to file
		err = WriteToFile(item.Path+"/BOOT.CFG", s)
		if err != nil {
			log.Fatal(err)
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

		if result := db.DB.Table("images").Create(&item); result.Error != nil {
			Error(c, http.StatusInternalServerError, result.Error) // 500
			return
		}
		c.JSON(http.StatusOK, item) // 200
	}
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

	//remove the entire directory and all files in it
	err = os.RemoveAll(item.Path)
	if err != nil {
		log.Fatal(err)
		Error(c, http.StatusInternalServerError, err) // 500
	}

	// remove record from database
	if res := db.DB.Delete(&item); res.Error != nil {
		Error(c, http.StatusInternalServerError, res.Error) // 500
		return
	}

	c.JSON(http.StatusNoContent, gin.H{}) //204
}

func WriteToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}
