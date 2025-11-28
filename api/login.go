package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"

	"github.com/sirupsen/logrus"
)

// Login handles user login
func Login(c *gin.Context) {

	var user models.UserLogin
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//get the user that is trying to authenticate

	var dbUser models.User
	if res := db.DB.Where("username = ?", user.Username).First(&dbUser); res.Error != nil {
		logrus.WithFields(logrus.Fields{
			"username": user.Username,
			"status":   "supplied username does not exist",
		}).Info("auth")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		Error(c, http.StatusUnauthorized, fmt.Errorf("supplied username does not exist")) // 404
		return
	}
	if ComparePasswords(dbUser.Password, []byte(user.Password), user.Username) {

		logrus.WithFields(logrus.Fields{
			"username": user.Username,
			"status":   "successfully authenticated",
		}).Debug("auth")
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		Error(c, http.StatusUnauthorized, fmt.Errorf("supplied username does not exist"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}
