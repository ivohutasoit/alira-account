package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ivohutasoit/alira-account/controller"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Println("$PORT must be set")
		port = "9000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	account := router.Group("/api/alpha/account")
	{
		account.GET("/register", controller.RegisterHandler)
	}

	auth := router.Group("/api/alpha/auth") 
	{
		auth.GET("/qrcode/:code", controller.GenerateImageQrcodeHandler)
		auth.POST("/socket/:code", controller.StartSocketHandler)
	}

	router.Run(":" + port)
}
