package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/common"
	"github.com/ivohutasoit/alira/util"
)

func RegisterHandler(c *gin.Context) {
	data := gin.H{
		"status": "OK",
	}
	c.JSONP(http.StatusOK, data)
}

func LoginPageHandler(c *gin.Context) {
	qrcode := &service.QrcodeService{}
	code := qrcode.Generate()
	
	encrypted, err := util.Encrypt(code, common.SecretKey)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	model.Tokens[code] = model.SocketLogin{
		Status: 1,
	}

	c.HTML(http.StatusOK, "login.tmpl.html", gin.H{
		"code": encrypted,
	})
}
