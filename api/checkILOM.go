package api

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/models"
)

// CheckIP handles the POST request for checking if the ip is available under given port
func CheckIP(c *gin.Context) {

	var ilo models.Ilo
	if err := c.ShouldBindJSON(&ilo); err != nil {
		log.Fatal(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try to reach the ip address using the given port
	_, err := net.DialTimeout("tcp", ilo.IloIpAddr+":"+ilo.Port, 2*time.Second)

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": fmt.Sprintf("IP %s is not reachable at port %v", ilo.IloIpAddr, ilo.Port)})
		return
	}

	// Example response, replace with actual logic
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("IP %s reached at port %v", ilo.IloIpAddr, ilo.Port)})
}
