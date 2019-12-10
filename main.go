package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira/model"
	"github.com/ivohutasoit/alira/model/domain"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func init() {
	model.GetDatabase().Debug().AutoMigrate(&domain.User{})
	model.GetDatabase().Debug().AutoMigrate(&domain.UserProfile{})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading .env file")
	}

	port := os.Getenv("PORT")

	if port == "" {
		fmt.Println("$PORT must be set")
		//port = "9000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.Default())

	store := cookie.NewStore([]byte(os.Getenv("SECRET_KEY")))
	router.Use(sessions.Sessions("ALIRASESSION", store))
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	web := router.Group("")
	{
		web.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, constant.IndexPage, nil)
		})
		webauth := web.Group("/auth")
		{
			webauth.GET("/login", controller.LoginPageHandler)
			webauth.POST("/login", controller.LoginPageHandler)
			webauth.GET("/logout", controller.LogoutPageHandler)
		}
	}

	api := router.Group("/api/alpha")
	{
		apiauth := api.Group("/auth")
		{
			apiauth.GET("/qrcode/:code", controller.GenerateImageQrcodeHandler)
			apiauth.GET("/socket/:code", controller.StartSocketHandler)
			apiauth.POST("/verify", controller.VerifyQrcodeHandler)
		}
	}

	router.Run(":" + port)
}
