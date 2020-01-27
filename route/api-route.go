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
		account := &controller.AccountController{}
		apiaccount := api.Group("/account")
		{
			apiaccount.POST("", account.CreateHandler)
			apiaccount.GET("/:id", account.DetailHandler)
			apiaccount.POST("/register", controller.RegisterHandler)
			apiaccount.POST("/profile", controller.ProfileHandler)
			apiaccount.POST("/identity", controller.IdentityHandler)
		}
		token := &controller.TokenController{}
		apitoken := api.Group("/token")
		{
			apitoken.POST("info", token.InfoHandler)
			apitoken.POST("verify", token.VerifyHandler)
		}
	}
}
