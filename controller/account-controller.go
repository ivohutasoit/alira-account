package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
)

func RegisterHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
		return
	}

	type Request struct {
		Reference string `form:"reference" json:"reference" xml:"reference" binding:"required"`
		UserID    string `form:"userid" json:"userid" xml:"userid" binding:"required"`
	}

	var request Request
	if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&request); err != nil {
			c.HTML(http.StatusBadRequest, constant.RegisterPage, nil)
			return
		}
	}

	c.Set("userid", request.UserID)

	switch request.Reference {
	case "mobile":
		RegisterByMobileHandler(c)
		return
	default:
		RegisterByEmailHandler(c)
		return
	}
}

// RegisterByEmailHandler for user registration using email address
func RegisterByEmailHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, nil)
	} else {
		accountService := &service.AccountService{}
		token, err := accountService.SendRegisterToken(c.GetString("userid"), "email")
		if err != nil {
			if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusBadRequest, gin.H{
					"code":   400,
					"status": "Bad Request",
					"error":  err.Error(),
				})
			} else {
				c.HTML(http.StatusOK, constant.RegisterPage, gin.H{
					"error": err.Error(),
				})
			}
			return
		}

		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
			c.JSONP(http.StatusOK, gin.H{
				"code":   200,
				"status": "OK",
				"data": map[string]string{
					"referer": token.Referer,
					"next":    "activate",
				},
			})
			return
		}
		c.HTML(http.StatusOK, constant.TokenPage, gin.H{
			"referer": token.Referer,
			"purpose": "ACTIVATION",
		})
	}
}

func RegisterByMobileHandler(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{
		"code":   200,
		"status": "OK",
	})
}

func CompleteProfileHandler(c *gin.Context) {
	type ProfileRequest struct {
		Username  string `form:"username" json:"username" xml:"username" binding:"required"`
		Email     string `form:"email" json:"email" xml:"email" binding:"required"`
		Mobile    string `form:"mobile" json:"mobile" xml:"mobile" binding:"required"`
		FirstName string `form:"first_name" json:"first_name" xml:"first_name" binding:"required"`
		LastName  string `form:"last_name" json:"last_name" xml:"last_name" binding:"required"`
		Gender    string `form:"gender" json:"gender" xml:"gender" binding:"gender"`
		Package   string `form:"package" json:"package" xml:"package" binding:"required"`
	}
	var request ProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":   400,
			"status": "Bad Request",
			"error":  err.Error(),
		})
		return
	}
}
