package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ivohutasoit/alira"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira-account/service"
	cstn "github.com/ivohutasoit/alira/constant"
	"github.com/ivohutasoit/alira/database/account"
	"github.com/ivohutasoit/alira/util"
	ua "github.com/mileusna/useragent"
	"github.com/skip2/go-qrcode"
)

type Auth struct{}

func (ctrl *Auth) LoginHandler(c *gin.Context) {
	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))

	redirect := c.Query("redirect")
	auth := &service.Auth{}
	socket, err := auth.GenerateLoginSocket(redirect)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	if c.Request.Method == http.MethodGet && !api {
		c.HTML(http.StatusOK, constant.LoginPage, gin.H{
			"code":     socket["code"].(string),
			"redirect": redirect,
		})
		return
	}
	type Request struct {
		AppID     string `form:"app_id" json:"app_id" xml:"app_id"`
		AppSecret string `form:"app_secret" json:"app_secret" xml:"app_secret"`
		UserID    string `form:"user_id" json:"user_id" xml:"user_id" binding:"required"`
	}
	var req Request
	if api {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.HTML(http.StatusUnauthorized, constant.LoginPage, gin.H{
				"code":     socket["code"].(string),
				"redirect": redirect,
				"error":    err.Error(),
			})
			return
		}
	}

	as := &service.Auth{}
	data, err := as.AuthenticateUser(req.UserID)
	if err != nil {
		if api {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":   http.StatusBadRequest,
				"status": http.StatusText(http.StatusBadRequest),
				"error":  err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, constant.LoginPage, gin.H{
			"code":     socket["code"].(string),
			"redirect": redirect,
			"error":    err.Error(),
		})
		return
	}
	user := data["user"].(*account.User)
	if api {
		if !user.UsePin {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"status":  http.StatusText(http.StatusOK),
				"message": "Login token has been sent to your email",
				"data": map[string]interface{}{
					"user_id": user.Model.ID,
					"purpose": data["purpose"].(string),
				},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"status":  http.StatusText(http.StatusOK),
			"message": "Please enter your pin",
			"data": map[string]interface{}{
				"user_id":      user.Model.ID,
				"pin_required": user.UsePin,
				"purpose":      data["purpose"].(string),
			},
		})
		return
	}
	if !user.UsePin {
		c.HTML(http.StatusOK, constant.TokenPage, gin.H{
			"redirect": redirect,
			"referer":  user.Model.ID,
			"purpose":  data["purpose"].(string),
		})
		return
	}

	c.HTML(http.StatusUnauthorized, constant.LoginPage, gin.H{
		"code":     socket["code"].(string),
		"redirect": redirect,
		"error":    "permission denied",
	})
}

// LoginHandler manages login request to show page, form action and api
func LoginHandler(c *gin.Context) {
	redirect := c.Query("redirect")

	auth := &service.Auth{}
	socket, err := auth.GenerateLoginSocket(redirect)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.LoginPage, gin.H{
			"code":     socket["code"].(string),
			"redirect": redirect,
		})
		return
	}

	type Request struct {
		UserID string `form:"userid" json:"userid" xml:"userid"  binding:"required"`
	}

	api := strings.Contains(c.Request.URL.Path, os.Getenv("URL_API"))
	var req Request
	if api {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.HTML(http.StatusUnauthorized, constant.IndexPage, gin.H{
				"code":     socket["code"].(string),
				"redirect": redirect,
				"error":    err.Error(),
			})
			return
		}
	}

	data, err := auth.SendLoginToken(req.UserID)
	if err != nil {
		if api {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
		c.HTML(http.StatusBadRequest, constant.LoginPage, gin.H{
			"code":     socket["code"].(string),
			"redirect": redirect,
			"error":    err.Error(),
		})
		return
	}
	status := data["status"].(string)

	if status == "SUCCESS" {
		if api {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"status":  "OK",
				"message": data["message"].(string),
				"data": map[string]string{
					"referer": data["referer"].(string),
				},
			})
			return
		}
		c.HTML(http.StatusOK, constant.TokenPage, gin.H{
			"referer":  data["referer"].(string),
			"redirect": redirect,
			"purpose":  data["purpose"].(string),
		})
		return
	}
}

func RefreshTokenHandler(c *gin.Context) {
	redirect := c.Query("redirect")
	currentPath := c.Request.URL.Path
	userid := c.MustGet("userid")
	tokens := strings.Split(c.Request.Header.Get("Authorization"), " ")

	Auth := &service.Auth{}
	data, err := Auth.GenerateRefreshToken(userid, tokens[1])
	if err != nil {
		fmt.Println(err.Error())
	}
	if strings.Contains(currentPath, os.Getenv("URL_API")) {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"code":   200,
			"status": "OK",
			"data": map[string]string{
				"access_token":  data["access_token"].(string),
				"refresh_token": data["refresh_token"].(string),
			},
		})
		return
	}

	session := sessions.Default(c)
	session.Set("access_token", data["access_token"].(string))
	session.Set("refresh_token", data["refresh_token"].(string))
	session.Save()

	if redirect != "" {
		uri, err := util.Decrypt(redirect, os.Getenv("SECRET_KEY"))
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
		}
		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s", uri))
		return
	}

	c.HTML(http.StatusOK, constant.IndexPage, nil)
}

func LogoutPageHandler(c *gin.Context) {
	Auth := &service.Auth{}
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		tokens := strings.Split(c.Request.Header.Get("Authorization"), " ")
		data, err := Auth.RemoveSessionToken(tokens[1])
		if err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}

		if data["status"].(string) == "success" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"status":  "OK",
				"message": data["message"].(string),
			})
			return
		}
	} else {
		redirect := c.Query("redirect")
		session := sessions.Default(c)
		_, err := Auth.RemoveSessionToken(session.Get("access_token"))
		if err != nil {
			fmt.Println(err.Error())
		}

		session.Delete("access_token")
		session.Delete("refresh_token")

		session.Save()

		if redirect != "" {
			uri, err := util.Decrypt(redirect, os.Getenv("SECRET_KEY"))
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
				return
			}
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s", uri))
			return
		}

		alira.ViewData = nil
		c.Redirect(http.StatusMovedPermanently, "/")
	}
}

var wsupgrader = &websocket.Upgrader{
	ReadBufferSize:  int(cstn.SocketBufferSize),
	WriteBufferSize: int(cstn.SocketBufferSize),
}

func GenerateImageQrcodeHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		var png []byte
		code := c.Query("code")

		c.Writer.Header().Set("Content-Type", "image/png")
		png, err := qrcode.Encode(code, qrcode.Medium, 256)
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
		}
		c.Writer.Write(png)
	}
}

func StartSocketHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		userAgent := c.Request.Header["User-Agent"][0]
		code := c.Query("code")
		decrypted, err := util.Decrypt(code, os.Getenv("SECRET_KEY"))
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
		}
		socket, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("Error while upgrading socket %v", err.Error())
			return
		}

		if model.Sockets[decrypted].Status != 1 {
			defer socket.Close()
			err := socket.WriteMessage(websocket.TextMessage, []byte("Token is not active"))
			if err != nil {
				return
			}
		}
		loginSocket := model.Sockets[decrypted]
		model.Sockets[decrypted] = model.LoginSocket{
			Redirect:  loginSocket.Redirect,
			UserAgent: userAgent,
			Socket:    socket,
		}
		for {
			mt, msg, err := socket.ReadMessage()
			if err != nil {
				fmt.Printf("Error while receiving message %v", err.Error())
				break
			}
			message := "Received " + string(msg)

			if err = socket.WriteMessage(mt, []byte(message)); err != nil {
				fmt.Printf("Error while sending message %v", err.Error())
				break
			}
		}
	}
}

func VerifyQrcodeHandler(c *gin.Context) {
	userAgent := ua.Parse(c.Request.Header["User-Agent"][0])

	if !userAgent.Mobile {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":   406,
			"status": "Not Acceptable",
			"error":  "qrcode log in must be used authenticated mobile app",
		})
		return
	}

	code := c.PostForm("code")
	decrypted, err := util.Decrypt(code, os.Getenv("SECRET_KEY"))
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	if model.Sockets[decrypted].Socket != nil {
		loginSocket := model.Sockets[decrypted]
		socket := loginSocket.Socket

		defer socket.Close()
		if loginSocket.Redirect != "" {
			uri, err := util.Decrypt(loginSocket.Redirect, os.Getenv("SECRET_KEY"))
			if err != nil {
				fmt.Println("Error: " + err.Error())
				return
			}
			err = socket.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s", uri)))
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
				return
			}
		} else {
			err = socket.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("http://%s", c.Request.Host)))
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
				return
			}
		}
		delete(model.Sockets, decrypted)
		model.Sockets[decrypted] = model.LoginSocket{
			Status: 0,
			Socket: socket,
		}

		//auth := &service.Auth{}
		//token, _ := auth.Login("Basic", "ivohutasoit", "hutasoit09")
		session := sessions.Default(c)
		//session.Set("access_token", token["access_token"])
		//session.Set("refresh_token", token["refresh_token"])
		session.Save()

		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"status":  "OK",
			"message": "token authentication successful",
		})
	}
}
