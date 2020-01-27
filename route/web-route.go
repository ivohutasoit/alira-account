package route

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/controller"
	"github.com/ivohutasoit/alira/middleware"
	"github.com/ivohutasoit/alira/model/domain"
)

type WebRoute struct{}

func (route *WebRoute) Initialize(r *gin.Engine) {
	web := r.Group("")
	web.Use(middleware.SessionHeaderRequired())
	{
		web.GET("/", func(c *gin.Context) {
			session := sessions.Default(c)
			session.Delete("message")
			session.Save()
			domain.Page["flash_message"] = session.Get("message")
			c.HTML(http.StatusOK, constant.IndexPage, domain.Page)
		})

		webauth := web.Group("/auth")
		{
			webauth.GET("/login", controller.LoginHandler)
			webauth.POST("/login", controller.LoginHandler)
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
