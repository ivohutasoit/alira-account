package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/messaging"
	"github.com/ivohutasoit/alira/util"
)

type Token struct{}

func (ctrl *Token) VerifyHandler(c *gin.Context) {
	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	if c.Request.Method == http.MethodPost {
		type Request struct {
			Referer      string `form:"referer" json:"referer" xml:"referer" binding:"required"`
			Token        string `form:"token" json:"token" xml:"token" binding:"required"`
			Purpose      string `form:"purpose" json:"purpose" xml:"purpose" binding:"required"`
			CustomerUser bool   `form:"customer_user" json:"customer_user" xml:"customer_user"`
		}
		var req Request
		if api {
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":   http.StatusBadRequest,
					"status": http.StatusText(http.StatusBadRequest),
					"error":  err.Error(),
				})
				return
			}
		} else {
			if err := c.ShouldBind(&req); err != nil {
				c.HTML(http.StatusBadRequest, constant.TokenPage, gin.H{
					"referer": req.Referer,
					"purpose": req.Purpose,
					"error":   err.Error(),
				})
				return
			}
		}

		var data map[interface{}]interface{}
		var err error
		if req.Purpose == "LOGIN" {

			authService := &service.Auth{}
			data, err = authService.VerifyToken(req.Referer, req.Token)
			if err != nil {
				c.HTML(http.StatusUnauthorized, constant.TokenPage, gin.H{
					"referer": req.Referer,
					"purpose": req.Purpose,
					"error":   err.Error(),
				})
				return
			}
			if api {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"status":  http.StatusText(http.StatusOK),
					"message": "Authentication successful",
					"data": map[string]string{
						"access_token":  data["access_token"].(string),
						"refresh_token": data["refresh_token"].(string),
					},
				})
				return
			}
			session := sessions.Default(c)
			session.Set("access_token", data["access_token"].(string))
			session.Set("refresh_token", data["refresh_token"].(string))
			session.Save()
			redirect := c.Query("redirect")
			var uri string
			if redirect != "" {
				uri, _ = util.Decrypt(redirect, os.Getenv("SECRET_KEY"))
				if strings.Contains(uri, "?") {
					uri = fmt.Sprintf("%s&callback=%s", uri, data["token_id"].(string))
				} else {
					uri = fmt.Sprintf("%s?callback=%s", uri, data["token_id"].(string))
				}
			} else {
				uri, _ = util.GenerateUrl(c.Request.TLS, c.Request.Host, "/", false)
			}

			if data["profile"].(string) == "required" {
				uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/account/profile?action=complete", false)
				c.Redirect(http.StatusMovedPermanently, uri)
				return
			}
			c.Redirect(http.StatusMovedPermanently, uri)
		} else if req.Purpose == "REGISTER" {
			as := &service.Account{}
			data, err = as.ActivateRegistration(req.Referer, req.Token)
			if err != nil {
				c.HTML(http.StatusUnauthorized, constant.TokenPage, gin.H{
					"referer": req.Referer,
					"purpose": req.Purpose,
					"error":   err.Error(),
				})
				return
			}
			if api {
				c.JSON(http.StatusOK, gin.H{
					"code":   http.StatusOK,
					"status": http.StatusText(http.StatusOK),
					"data": map[string]string{
						"userid":        data["userid"].(string),
						"access_token":  data["access_token"].(string),
						"refresh_token": data["refresh_token"].(string),
						"profile":       "required",
					},
				})
				return
			}
			session := sessions.Default(c)
			session.Set("access_token", data["access_token"].(string))
			session.Set("refresh_token", data["refresh_token"].(string))
			session.Set("required_profile", true)
			session.Save()

			uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/account/profile?action=complete", false)
			c.Redirect(http.StatusMovedPermanently, uri)
		} else {

		}
	}
}

func (ctrl *Token) CallbackHandler(c *gin.Context) {
	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	type Request struct {
		AppID     string `form:"app_id" json:"app_id" xml:"app_id"`
		AppSecret string `form:"app_secret" json:"app_secret" xml:"app_secret"`
		Token     string `form:"token" json:"token" xml:"token" binding:"required"`
	}
	var req Request
	if api {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
			return
		}
	}

	tokenService := &service.Token{}
	data, err := tokenService.Get(req.Token)
	if err != nil {
		if api {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
			return
		}
	}

	if api {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"status":  http.StatusText(http.StatusOK),
			"message": "Authenticated session token",
			"data": map[string]string{
				"access_token":  data["access_token"].(string),
				"refresh_token": data["refresh_token"].(string),
			},
		})
	}
}

func (ctrl *Token) InfoHandler(c *gin.Context) {
	// 1. Check whitelist url
	/*data := os.Getenv("IP_WHITELIST")
	if data == "" {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"status":  http.StatusText(http.StatusForbidden),
				"message": "Permission denied",
			})
			return
		}
	}

	hasAccess := false
	whitelist := strings.Split(data, ";")
	for _, item := range whitelist {
		if item == c.ClientIP() {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"status":  http.StatusText(http.StatusForbidden),
				"message": "Permission denied",
			})
			return
		}
	}*/

	type Request struct {
		AppID     string `form:"app_id" json:"app_id" xml:"app_id"`
		AppSecret string `form:"app_secret" json:"app_secret" xml:"app_secret"`
		Type      string `form:"type" json:"type" xml:"type" binding:"required"`
		Token     string `form:"token" json:"token" xml:"token" binding:"required"`
	}
	var req Request
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
			return
		}
	}

	tokenService := &service.Token{}
	data, err := tokenService.GetAuthenticated(req.Type, req.Token)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
			return
		}
	}

	user := data["user"].(*account.User)
	userProfile := &messaging.UserProfile{
		ID:            user.Model.ID,
		Username:      user.Username,
		Email:         user.Email,
		PrimaryMobile: user.Mobile,
		Active:        user.Active,
		Avatar:        user.Avatar,
	}

	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"status":  http.StatusText(http.StatusOK),
			"message": "Authenticated user",
			"data":    userProfile,
		})
	}
}
