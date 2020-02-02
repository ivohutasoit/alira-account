package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira-account/middleware"
)

type WebRoute struct{}

func (route *WebRoute) Initialize(r *gin.Engine) {
	authMiddleware := &middleware.Auth{}
	web := r.Group("")
	web.Use(authMiddleware.SessionRequired())
	{
		indexController := &controller.Index{}
		web.GET("", indexController.IndexHandler)

		webauth := web.Group("/auth")
		{
			auth := &controller.Auth{}
			webauth.GET("/login", auth.LoginHandler)
			webauth.POST("/login", auth.LoginHandler)
			webauth.GET("/qrcode", controller.GenerateImageQrcodeHandler)
			webauth.GET("/socket", controller.StartSocketHandler)
			webauth.GET("/logout", controller.LogoutPageHandler)
		}
		webacct := web.Group("/account")
		{
			webacct.GET("/register", controller.RegisterHandler)
			webacct.POST("/register", controller.RegisterHandler)
			webacct.GET("/profile", controller.ProfileHandler)
			webacct.POST("/profile", controller.ProfileHandler)
		}
		tokenController := &controller.TokenController{}
		webtoken := web.Group("/token")
		{
			webtoken.POST("/verify", tokenController.VerifyHandler)
		}
	}
}
