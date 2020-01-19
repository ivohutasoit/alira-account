package controller

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/util"
)

type TokenController struct{}

func (ctrl *TokenController) VerifyHandler(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		type Request struct {
			Referer string `form:"referer" json:"referer" xml:"referer" binding:"required"`
			Token   string `form:"token" json:"token" xml:"token" binding:"required"`
			Purpose string `form:"purpose" json:"purpose" xml:"purpose" binding:"required"`
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
				c.HTML(http.StatusUnauthorized, constant.TokenPage, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
		var data map[interface{}]interface{}
		var err error
		if req.Purpose == "LOGIN" {
			authService := &service.AuthService{}
			data, err = authService.VerifyLoginToken(req.Referer, req.Token)
			if err != nil {
				c.HTML(http.StatusUnauthorized, constant.TokenPage, gin.H{
					"referer": req.Referer,
					"purpose": req.Purpose,
					"error":   err.Error(),
				})
				return
			}
			if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusOK, gin.H{
					"code":   200,
					"status": "OK",
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
			} else {
				uri, _ = util.GenerateUrl(c.Request.TLS, c.Request.Host, "/", false)
			}
			c.Redirect(http.StatusMovedPermanently, uri)
		} else if req.Purpose == "REGISTER" {
			accService := &service.AccountService{}
			data, err = accService.ActivateRegistration(req.Referer, req.Token)
			if err != nil {
				c.HTML(http.StatusUnauthorized, constant.TokenPage, gin.H{
					"referer": req.Referer,
					"purpose": req.Purpose,
					"error":   err.Error(),
				})
				return
			}
			if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusOK, gin.H{
					"code":   200,
					"status": "OK",
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
		}
	}
}

func (ctrl *TokenController) InfoHandler(c *gin.Context) {
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
		Type  string `form:"type" json:"type" xml:"type" binding:"required"`
		Token string `form:"token" json:"token" xml:"token" binding:"required"`
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

	tokenService := &service.TokenService{}
	data, err := tokenService.GetTokenInformation(req.Type, req.Token)
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

	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"status":  http.StatusText(http.StatusOK),
			"message": "Authenticated user",
			"data": map[string]string{
				"userid": data["userid"].(string),
				"client": c.ClientIP(),
				"host":   c.Request.URL.Host,
			},
		})
	}
}
