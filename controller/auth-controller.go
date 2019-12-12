package controller

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/common"
	"github.com/ivohutasoit/alira/util"
	ua "github.com/mileusna/useragent"
	"github.com/skip2/go-qrcode"
)

// LoginHandler manages login request to show page, form action and api
func LoginHandler(c *gin.Context) {
	redirect := c.Query("redirect")

	qrcode := &service.QrcodeService{}
	code := qrcode.Generate()

	encrypted, err := util.Encrypt(code, os.Getenv("SECRET_KEY"))
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	model.Sockets[code] = model.LoginSocket{
		Redirect: redirect,
		Status:   1,
	}

	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.LoginPage, gin.H{
			"code":     encrypted,
			"redirect": redirect,
		})
	} else {
		type Login struct {
			UserID   string `form:"userid" json:"userid" xml:"userid"  binding:"required"`
			Password string `form:"password" json:"password" xml:"password" binding:"required"`
		}
		var login Login
		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
			if err := c.ShouldBindJSON(&login); err != nil {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusBadRequest, gin.H{
					"code":   400,
					"status": "Bad Request",
					"error":  err.Error(),
				})
				return
			}
		} else {
			if err := c.ShouldBind(&login); err != nil {
				c.HTML(http.StatusUnauthorized, constant.IndexPage, gin.H{
					"code":     encrypted,
					"redirect": redirect,
					"error":    err.Error(),
				})
				return
			}
		}

		fmt.Println(c.PostForm("userid"))
		auth := &service.AuthService{}
		token, err := auth.Login("Basic", login.UserID, login.Password)
		if err != nil {
			c.HTML(http.StatusBadRequest, constant.IndexPage, gin.H{
				"code":     encrypted,
				"redirect": redirect,
				"error":    err.Error(),
			})
			return
		}

		if strings.Contains(c.Request.URL.Path, os.Getenv("API_URI")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"code":   200,
				"status": "OK",
				"data": map[string]string{
					"access_token":  token["access_token"].(string),
					"refresh_token": token["refresh_token"].(string),
				},
			})
			return
		}

		session := sessions.Default(c)
		session.Set("access_token", token["access_token"])
		session.Set("refresh_token", token["refresh_token"])
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
}

func RefreshTokenHandler(c *gin.Context) {
	redirect := c.Query("redirect")
	currentPath := c.Request.URL.Path

	userid := c.MustGet("userid")

	auth := &service.AuthService{}
	token, err := auth.Login("Refresh", userid)
	if err != nil {
		fmt.Println(err.Error())
	}
	if currentPath == fmt.Sprintf("%s%s", os.Getenv("API_URI"), "/auth/refresh") {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"code":   200,
			"status": "OK",
			"data": map[string]string{
				"access_token":  token["access_token"].(string),
				"refresh_token": token["refresh_token"].(string),
			},
		})
		return
	}

	session := sessions.Default(c)
	session.Set("access_token", token["access_token"])
	session.Set("refresh_token", token["refresh_token"])
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
	redirect := c.Query("redirect")
	session := sessions.Default(c)
	session.Clear()
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

	c.HTML(http.StatusOK, constant.IndexPage, nil)
}

var wsupgrader = &websocket.Upgrader{
	ReadBufferSize:  int(common.SocketBufferSize),
	WriteBufferSize: int(common.SocketBufferSize),
}

func GenerateImageQrcodeHandler(c *gin.Context) {
	var png []byte
	code := c.Param("code")

	c.Writer.Header().Set("Content-Type", "image/png")
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	c.Writer.Write(png)
}

func StartSocketHandler(c *gin.Context) {
	userAgent := c.Request.Header["User-Agent"][0]
	code := c.Param("code")
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

		auth := &service.AuthService{}
		token, _ := auth.Login("Basic", "ivohutasoit", "hutasoit09")
		session := sessions.Default(c)
		session.Set("access_token", token["access_token"])
		session.Set("refresh_token", token["refresh_token"])
		session.Save()

		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"status":  "OK",
			"message": "token authentication successful",
		})
	}
}
