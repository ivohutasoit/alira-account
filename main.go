package main

import (
	"fmt"
	"os"

	"github.com/ivohutasoit/alira-account/route"
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

	web := &route.WebRoute{}
	web.Initialize(router)

	api := &route.ApiRoute{}
	api.Initialize(router)

	router.Run(":" + port)
}
