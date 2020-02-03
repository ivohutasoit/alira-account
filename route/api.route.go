package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira-account/middleware"
)

type API struct{}

func (route *API) Initialize(r *gin.Engine) {
	authMiddleware := &middleware.Auth{}
	api := r.Group("/api/alpha")
	api.Use(authMiddleware.TokenRequired())
	{
		auth := &controller.Auth{}
		apiauth := api.Group("/auth")
		{
			apiauth.POST("/login", auth.LoginHandler)
			apiauth.POST("/refresh", controller.RefreshTokenHandler)
			apiauth.POST("/verify", controller.VerifyQrcodeHandler)
			apiauth.POST("/logout", controller.LogoutPageHandler)
		}
		ac := &controller.Account{}
		apiaccount := api.Group("/account")
		{
			apiaccount.POST("", ac.CreateHandler)
			apiaccount.GET("/:id", ac.DetailHandler)
			apiaccount.POST("/pin", ac.ChangePinHandler)
			apiaccount.POST("/register", controller.RegisterHandler)
			apiaccount.POST("/profile", controller.ProfileHandler)
			apiaccount.POST("/identity", controller.IdentityHandler)
		}
		token := &controller.Token{}
		apitoken := api.Group("/token")
		{
			apitoken.POST("/callback", token.CallbackHandler)
			apitoken.POST("/info", token.InfoHandler)
			apitoken.POST("/verify", token.VerifyHandler)
		}
	}
}
