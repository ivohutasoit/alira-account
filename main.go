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

	web := router.Group("/") 
	{
		web.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.tmpl.html", nil)
		})
		web.GET("/login", controller.LoginPageHandler)
	}

	api := router.Group("/api/alpha")
	{
		apiauth := api.Group("/auth")
		{
			apiauth.GET("/qrcode/:code", controller.GenerateImageQrcodeHandler)
			apiauth.GET("/socket/:code", controller.StartSocketHandler)
			apiauth.GET("/verify/:code", controller.VerifyQrcodeHandler)
		}
	}

	router.Run(":" + port)
}
