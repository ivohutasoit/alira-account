package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
		tokenController := &controller.TokenController{}
		webtoken := web.Group("/token")
		{
			webtoken.POST("/verify", tokenController.VerifyHandler)
		}
	}
}
