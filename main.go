package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira/middleware"
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
	model.GetDatabase().Debug().AutoMigrate(&domain.User{},
		&domain.Profile{},
		&domain.Subscribe{},
		&domain.Token{})
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
	router.LoadHTMLGlob("views/*/*.tmpl.html")
	router.Static("/static", "static")

	web := router.Group("")
	{
		web.Use(middleware.SessionHeaderRequired(os.Getenv("LOGIN_URL")))
		web.Any("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, constant.IndexPage, gin.H{
				"userid": c.GetString("userid"),
			})
		})
		webauth := web.Group("/auth")
		{
			webauth.GET("/login", controller.LoginHandler)
			webauth.POST("/login", controller.LoginHandler)
			webauth.GET("/qrcode/:code", controller.GenerateImageQrcodeHandler)
			webauth.GET("/socket/:code", controller.StartSocketHandler)
			webauth.GET("/logout", controller.LogoutPageHandler)
		}
		webacct := web.Group("/account")
		{
			webacct.GET("/register", controller.RegisterHandler)
			webacct.POST("/register", controller.RegisterHandler)
		}
		webtoken := web.Group("/token")
		{
			webtoken.POST("/verify", controller.VerifyTokenHandler)
		}
	}

	api := router.Group("/api/alpha")
	api.Use(middleware.TokenHeaderRequired())
	{
		apiauth := api.Group("/auth")
		{
			apiauth.POST("/login", controller.LoginHandler)
			apiauth.POST("/refresh", controller.RefreshTokenHandler)
			apiauth.POST("/verify", controller.VerifyQrcodeHandler)
		}
		apiaccount := api.Group("/account")
		{
			//apiaccount.GET("/:id", controller.ProfileHandler)
			apiaccount.POST("/register", controller.RegisterHandler)
			apiaccount.POST("/profile", controller.CompleteProfileHandler)
		}
		apitoken := api.Group("token")
		{
			apitoken.POST("verify", controller.VerifyTokenHandler)
		}
	}

	router.Run(":" + port)
}
