package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/ivohutasoit/alira/util"
)

func RegisterHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.RegisterPage, c.Request.TLS.NegotiatedProtocolIsMutual)
		return
	}

	type Request struct {
		Using   string `form:"using" json:"using" xml:"using" binding:"required"`
		Payload string `form:"payload" json:"payload" xml:"payload" binding:"required"`
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

	c.Set("payload", request.Payload)

	switch request.Using {
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
		token, err := accountService.SendRegisterToken(c.GetString("payload"), "email")
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

func ActivateHandler(c *gin.Context) {
	type Request struct {
		Referer string `form:"referer" json:"referer" xml:"referer" binding:"required"`
		Token   string `form:"token" json:"token" xml:"token" binding:"required"`
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

	acctService := &service.AccountService{}
	result, err := acctService.ActivateRegistration(request.Referer, request.Token)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	}

	if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
		c.Header("Content-Type", "application/json")
		c.JSONP(http.StatusOK, gin.H{
			"code":    200,
			"status":  "OK",
			"message": "your user account created successfull",
			"data": map[string]interface{}{
				"access_token":     result["access_token"],
				"refresh_token":    result["refresh_token"],
				"profile_required": true,
			},
		})
		return
	}

	session := sessions.Default(c)
	session.Set("access_token", result["access_token"])
	session.Set("refresh_token", result["refresh_token"])
	session.Set("profile_required", true)
	session.Save()

	uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/", false)
	c.Redirect(http.StatusPermanentRedirect, uri)
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

	user := &domain.User{
		Username: request.Username,
		Email:    request.Email,
		Mobile:   request.Mobile,
		Active:   true,
	}
	model.GetDatabase().Create(user)

	profile := &domain.Profile{
		BaseModel: model.BaseModel{
			ID: user.BaseModel.ID,
		},
		User:      *user,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Gender:    request.Gender,
	}
	model.GetDatabase().Create(profile)

	subscribe := &domain.Subscribe{
		UserID:    user.BaseModel.ID,
		Code:      request.Package,
		User:      *user,
		Purpose:   "USEAPP",
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(2, 0, 0),
		AgreedAt:  time.Now(),
	}
	model.GetDatabase().Create(subscribe)
	auth := &service.AuthService{}
	token, err := auth.Login("Refresh", user.BaseModel.ID)
	if err != nil {
		fmt.Println(err.Error())
	}

	/*login := &domain.Session{
		UserID:       user.BaseModel.ID,
		User:         user,
		Agent:        c.Request.Header["User-Agent"][0],
		AccessToken:  token["access_token"].(string),
		RefreshToken: token["refresh_token"].(string),
		NotBefore:    time.Now(),
	}
	model.GetDatabase().Create(login)

	if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusCreated, gin.H{
			"code":   201,
			"status": "Created",
			"data": map[string]string{
				"access_token":  login.AccessToken,
				"refresh_token": login.RefreshToken,
			},
		})
		return
	}*/

	session := sessions.Default(c)
	session.Set("access_token", token["access_token"])
	session.Set("refresh_token", token["refresh_token"])
	session.Save()

	fmt.Println(token["access_token"])

	c.HTML(http.StatusOK, constant.IndexPage, gin.H{
		"flash": "your user account created successfull",
	})
}
