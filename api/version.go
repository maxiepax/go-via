package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Version(version string, commit string, date string) func(c *gin.Context) {
	return func(c *gin.Context) {

		type Version struct {
			Version string
			Commit  string
			Date    string
		}

		item := Version{version, commit, date}

		c.JSON(http.StatusOK, item) // 200
	}
}
