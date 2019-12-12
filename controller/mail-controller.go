package controller

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gin-gonic/gin"
)

func SendMailHandler(c *gin.Context) {
	err := smtp.SendMail(fmt.Sprintf("%s:%s", os.Getenv("SMTP.HOST"), os.Getenv("SMTP.PORT")),
		smtp.PlainAuth("", os.Getenv("SMTP.SENDER"), os.Getenv("SMTP.PASSWORD"), os.Getenv("SMTP.HOST")),
		os.Getenv("SMTP.SENDER"), []string{"if09051@gmail.com"}, []byte(
			"From: "+os.Getenv("SMTP.SENDER")+"\n"+
				"To: if09051@gmail.com\n"+
				"Subject: Hello there\n"))
	if err != nil {
		fmt.Printf("smtp error: %s", err.Error())
		return
	}
	c.JSONP(http.StatusOK, nil)
}
