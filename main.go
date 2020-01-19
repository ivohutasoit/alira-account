package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ivohutasoit/alira/model/domain"

	"github.com/joho/godotenv"

	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira/middleware"
	"github.com/joho/godotenv"

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

	tokenController := &controller.TokenController{}

	web := router.Group("")
	{
		web.Use(middleware.SessionHeaderRequired(os.Getenv("URL_LOGIN")))
		web.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			flashMessage := session.Get("message")
			session.Delete("message")
			session.Save()
			data := map[string]string{
				"type":  "Bearer",
				"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NzkyNjYzMDgsImp0aSI6IjA5ODZlOWYxLTliYjctNDEwOS1iYTI2LWE0MjBiNmJhYTE4YiIsImlhdCI6MTU3OTE3OTkwOCwibmJmIjoxNTc5MTc5OTA4LCJVc2VySUQiOiIiLCJBZG1pbiI6ZmFsc2V9.tnN-Rt56qOn4RU1opGEHOVFp-ZAxRNH8muKRnZ-ivY4",
			}
			// https://tutorialedge.net/golang/consuming-restful-api-with-go/
			payload, _ := json.Marshal(data)
			resp, err := http.Post(os.Getenv("URL_AUTH"), "application/json", bytes.NewBuffer(payload))
			if err != nil {
				fmt.Println(err.Error())
			}
			respData, _ := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err.Error())
			}
			var response domain.Response
			json.Unmarshal(respData, &response)
			c.HTML(http.StatusOK, constant.IndexPage, gin.H{
				"userid":        c.GetString("userid"),
				"flash_message": flashMessage,
				"data":          response.Data["userid"],
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
			apitoken.POST("info", tokenController.InfoHandler)
			apitoken.POST("verify", tokenController.VerifyHandler)
		}
	}

	router.Run(":" + port)
}
