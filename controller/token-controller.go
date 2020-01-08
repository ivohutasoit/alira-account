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

func VerifyTokenHandler(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		type Request struct {
			Referer string `form:"referer" json:"referer" xml:"referer" binding:"required"`
			Token   string `form:"token" json:"token" xml:"token" binding:"required"`
			Purpose string `form:"purpose" json:"purpose" xml:"purpose" binding:"required"`
		}
		var req Request
		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
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
					"purpose": "LOGIN",
					"error":   err.Error(),
				})
				return
			}
			if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
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

			uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/", false)
			c.Redirect(http.StatusPermanentRedirect, uri)
		} else if req.Purpose == "ACTIVATION" {
			accService := &service.AccountService{}
			data, err = accService.ActivateRegistration(req.Referer, req.Token)
			if err != nil {
				c.HTML(http.StatusUnauthorized, constant.TokenPage, gin.H{
					"referer": req.Referer,
					"purpose": "ACTIVATION",
					"error":   err.Error(),
				})
				return
			}
			if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
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

			uri, _ := util.GenerateUrl(c.Request.TLS, c.Request.Host, "/", false)
			c.Redirect(http.StatusPermanentRedirect, uri)
		}
	}
}
