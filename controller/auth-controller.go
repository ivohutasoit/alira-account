package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira-account/service"
	"github.com/ivohutasoit/alira/common"
	"github.com/ivohutasoit/alira/util"
	"github.com/skip2/go-qrcode"
)

func LoginPageHandler(c *gin.Context) {
	callback := c.Query("callback")

	qrcode := &service.QrcodeService{}
	code := qrcode.Generate()

	encrypted, err := util.Encrypt(code, os.Getenv("SECRET_KEY"))
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	model.Sockets[code] = model.LoginSocket{
		Callback: callback,
		Status:   1,
	}

	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, "login.tmpl.html", gin.H{
			"code":     encrypted,
			"callback": callback,
		})
	} else {
		token, err := service.Login(c.PostForm("userid"), c.PostForm("password"))
		if err != nil {
			c.HTML(http.StatusUnauthorized, "login.tmpl.html", gin.H{
				"code":     encrypted,
				"callback": callback,
				"error":    err.Error(),
			})
			return
		}

		session := sessions.Default(c)
		session.Set("token", token)
		session.Save()

		if callback != "" {
			uri, err := util.Decrypt(callback, os.Getenv("SECRET_KEY"))
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
			}
			c.Redirect(http.StatusMovedPermanently, uri)
			return
		}

		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	}
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
	model.Sockets[decrypted] = model.LoginSocket{
		Socket: socket,
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
	code := c.Param("code")
	decrypted, err := util.Decrypt(code, os.Getenv("SECRET_KEY"))
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	if model.Sockets[decrypted].Socket != nil {
		loginSocket := model.Sockets[decrypted]
		socket := loginSocket.Socket

		defer socket.Close()
		if loginSocket.Callback != "" {
			uri, err := util.Decrypt(loginSocket.Callback, os.Getenv("SECRET_KEY"))
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
			}
			err = socket.WriteMessage(websocket.TextMessage, []byte(uri))
			if err != nil {
				return
			}
		} else {
			err = socket.WriteMessage(websocket.TextMessage, []byte("http://localhost:9000"))
			if err != nil {
				return
			}
		}
		delete(model.Sockets, decrypted)
		model.Sockets[decrypted] = model.LoginSocket{
			Status: 0,
			Socket: socket,
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}
