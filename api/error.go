package api

import (
	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/models"
)

func Error(c *gin.Context, status int, err error) {
	c.JSON(status, models.APIError{
		ErrorStatus:  status,
		ErrorMessage: err.Error(),
	})
}
