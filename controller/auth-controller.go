package controller

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira/common"
	"github.com/ivohutasoit/alira/util"
	"github.com/skip2/go-qrcode"
)

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
	fmt.Println("Start Socket")
	code := c.Param("code")
	decrypted, err := util.Decrypt(code, common.SecretKey)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	socket, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error while upgrading socket %v", err.Error())
		return
	}

	if model.Tokens[decrypted].Status != 1 {
		defer socket.Close()
		err := socket.WriteMessage(websocket.TextMessage, []byte("Token is not active"))
		if err != nil {
			return
		}
	}
	model.Tokens[decrypted] = model.SocketLogin{
		Socket: socket,
	}
	fmt.Println(model.Tokens[decrypted])
	for {
		mt, msg, err := socket.ReadMessage()
		if err != nil {
			log.Printf("Error while receiving message %v", err.Error())
			break
		}
		message := "Received " + string(msg)

		if err = socket.WriteMessage(mt, []byte(message)); err != nil {
			log.Printf("Error while sending message %v", err.Error())
			break
		}
	}
}

func VerifyQrcodeHandler(c *gin.Context) {
	fmt.Println("Verify Qrcode")
	code := c.Param("code")
	decrypted, err := util.Decrypt(code, common.SecretKey)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	if  model.Tokens[decrypted].Socket != nil {
		socket :=  model.Tokens[decrypted].Socket

		defer socket.Close()
		err := socket.WriteMessage(websocket.TextMessage, []byte("Token is valid"))
		if err != nil {
			return
		}
		delete(model.Tokens, decrypted)
		model.Tokens[decrypted] = model.SocketLogin{
			Status: 0,
			Socket: socket,
		}
	}
}
