package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/util"
)

type Index struct{}

func (ctrl *Index) IndexHandler(c *gin.Context) {
	source := c.Query("source")
	redirect := c.Query("redirect")
	session := sessions.Default(c)
	if source != "" && source == "logout" {
		authService := &service.AuthService{}
		_, err := authService.RemoveSessionToken(session.Get("access_token"))
		if err != nil {
			fmt.Println(err.Error())
		}
		session.Clear()
	}
	if redirect != "" {
		uri, err := util.Decrypt(redirect, os.Getenv("SECRET_KEY"))
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
			return
		}
		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s", uri))
		return
	}
	if alira.ViewData == nil {
		alira.ViewData = gin.H{
			"flash_message": session.Get("message"),
		}
	} else {
		alira.ViewData["flash_message"] = session.Get("message")
	}
	session.Delete("message")
	session.Save()
	c.HTML(http.StatusOK, constant.IndexPage, alira.ViewData)
}
