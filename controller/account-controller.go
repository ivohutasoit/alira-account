package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/util"
)

func AccountViewHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		return
	}
}

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
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
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
			c.HTML(http.StatusBadRequest, constant.RegisterPage, gin.H{
				"error": err.Error(),
			})
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
		return
	}

	accountService := &service.AccountService{}
	data, err := accountService.SendRegisterToken(c.GetString("userid"), "email")
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
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

	if data["status"].(string) == "SUCCESS" {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.JSONP(http.StatusOK, gin.H{
				"code":   200,
				"status": "OK",
				"data": map[string]string{
					"referer": data["referer"].(string),
					"next":    "activate",
				},
			})
			return
		}

		c.HTML(http.StatusOK, constant.TokenPage, gin.H{
			"referer": data["referer"].(string),
			"purpose": data["purpose"].(string),
			"message": data["message"].(string),
		})
	}
}

func RegisterByMobileHandler(c *gin.Context) {
	c.JSONP(http.StatusOK, gin.H{
		"code":   200,
		"status": "OK",
	})
}

func ProfileHandler(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		action = "view"
	}
	accService := &service.AccountService{}
	data, err := accService.Get(c.GetString("userid"))
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
		} else {
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"userid": c.GetString("userid"),
				"error":  err.Error(),
			})
		}
		return
	}

	user := data["user"].(*domain.User)
	profile := data["profile"].(*domain.Profile)
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.ProfilePage, gin.H{
			"userid":     user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"mobile":     user.Mobile,
			"first_name": profile.FirstName,
			"last_name":  profile.LastName,
			"gender":     strings.ToLower(profile.Gender),
			"state":      action,
		})
		return
	}
	type Request struct {
		UserID    string `form:"userid" json:"userid" xml:"userid" binding:"required"`
		Username  string `form:"username" json:"username" xml:"username" binding:"required"`
		Mobile    string `form:"mobile" json:"mobile" xml:"mobile" binding:"required"`
		FirstName string `form:"first_name" json:"first_name" xml:"first_name" binding:"required"`
		LastName  string `form:"last_name" json:"last_name" xml:"last_name" binding:"required"`
		Gender    string `form:"gender" json:"gender" xml:"gender" binding:"required"`
	}
	var req Request
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"userid":     user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"mobile":     user.Mobile,
				"first_name": profile.FirstName,
				"last_name":  profile.LastName,
				"gender":     strings.ToLower(profile.Gender),
				"state":      action,
				"error":      err.Error(),
			})
			return
		}
	}
	data, err = accService.SaveProfile(req.UserID, req.Username, req.Mobile, req.FirstName, req.LastName, req.Gender)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
		} else {
			c.HTML(http.StatusBadRequest, constant.ProfilePage, gin.H{
				"userid":     user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"mobile":     user.Mobile,
				"first_name": profile.FirstName,
				"last_name":  profile.LastName,
				"gender":     strings.ToLower(profile.Gender),
				"state":      action,
				"error":      err.Error(),
			})
		}
		return
	}

	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"status":  "OK",
			"message": data["message"].(string),
			"data": map[string]string{
				"userid": data["userid"].(string),
			},
		})
		return
	}

	session := sessions.Default(c)
	session.Set("message", data["message"])
	session.Save()
	uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/account/profile", false)
	c.Redirect(http.StatusMovedPermanently, uri)
}
