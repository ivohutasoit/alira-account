package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira/middleware"
)

type ApiRoute struct{}

func (route *ApiRoute) Initialize(r *gin.Engine) {
	api := r.Group("/api/alpha")
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
		tokenController := &controller.TokenController{}
		apitoken := api.Group("/token")
		{
			apitoken.POST("info", tokenController.InfoHandler)
			apitoken.POST("verify", tokenController.VerifyHandler)
		}
	}
}
