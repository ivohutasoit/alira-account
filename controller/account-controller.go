package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
)

// RegisterByEmailHandler for user registration using email address
func RegisterByEmailHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
	} else {
		type Registration struct {
			Email string `form:"email" json:"email" xml:"email"  binding:"required"`
		}
		var registration Registration
		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
			if err := c.ShouldBindJSON(&registration); err != nil {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusBadRequest, gin.H{
					"code":   400,
					"status": "Bad Request",
					"error":  err.Error(),
				})
				return
			}
		} else {
			if err := c.ShouldBind(&registration); err != nil {
				c.HTML(http.StatusBadRequest, constant.RegisterPage, nil)
				return
			}
		}
		data := gin.H{
			"status": "OK",
		}
		c.JSONP(http.StatusOK, data)
	}
}
