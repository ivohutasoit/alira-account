package main

import (
	"fmt"
	"os"
	"time"

	alira "github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira-account/route"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func init() {
	alira.GetConnection().Debug().AutoMigrate(&account.User{},
		&account.Profile{},
		&account.Subscription{},
		&account.Token{},
		&account.Identity{},
		&account.NationalIdentity{})
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
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			fmt.Println(origin)
			return true
		},
	}))

	store := cookie.NewStore([]byte(os.Getenv("SECRET_KEY")))
	router.Use(sessions.Sessions("ALIRASESSION", store))
	router.LoadHTMLGlob("views/*/*.tmpl.html")
	router.Static("/static", "static")

	web := &route.Web{}
	web.Initialize(router)

	api := &route.API{}
	api.Initialize(router)

	router.Run(":" + port)
}
