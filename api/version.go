package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Version(commit string, date string) func(c *gin.Context) {
	return func(c *gin.Context) {

		type Version struct {
			Commit string
			Date   string
		}

		item := Version{commit, date}

		c.JSON(http.StatusOK, item) // 200
	}
}
