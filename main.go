package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

/*func init() {
	model.GetDatabase().Debug().AutoMigrate(&domain.User{},
		&domain.Profile{},
		&domain.Subscribe{},
		&domain.Token{},
		&domain.Identity{},
		&domain.NationalIdentity{})
}*/

func main() {
	//err := godotenv.Load()
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

	tokenController := &controller.TokenController{}

	web := router.Group("")
	{
		web.Use(middleware.SessionHeaderRequired(os.Getenv("URL_LOGIN")))
		web.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			flashMessage := session.Get("message")
			session.Delete("message")
			session.Save()
			response, err := http.Get("http://localhost:9000/api/alpha/token")
			if err != nil {
				c.HTML(http.StatusOK, constant.IndexPage, gin.H{
					"userid":        c.GetString("userid"),
					"flash_message": flashMessage,
					"error":         err.Error(),
				})
			}
			data, _ := ioutil.ReadAll(response.Body)
			c.HTML(http.StatusOK, constant.IndexPage, gin.H{
				"userid":        c.GetString("userid"),
				"flash_message": flashMessage,
				"data":          string(data),
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
			webacct.GET("/profile", controller.ProfileHandler)
			webacct.POST("/profile", controller.ProfileHandler)
		}
		webtoken := web.Group("/token")
		{
			webtoken.POST("/verify", tokenController.VerifyHandler)
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
			apiauth.POST("/logout", controller.LogoutPageHandler)
		}
		apiaccount := api.Group("/account")
		{
			apiaccount.POST("/register", controller.RegisterHandler)
			apiaccount.POST("/profile", controller.ProfileHandler)
			apiaccount.POST("/identity", controller.IdentityHandler)
		}
		apitoken := api.Group("/token")
		{
			apitoken.GET("", tokenController.DetailHandler)
			apitoken.POST("verify", tokenController.VerifyHandler)
		}
	}

	router.Run(":" + port)
}
