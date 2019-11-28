package controller

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ivohutasoit/alira-account/model"
	"github.com/ivohutasoit/alira/common"
	"github.com/ivohutasoit/alira/util"
	"github.com/skip2/go-qrcode"
)

var tokens = make(map[string]model.SocketLogin)

var wsupgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func GenerateQrcodeHandler() (token string) {
	b := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	encrypted, err := util.Encrypt(string(b))
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	tokens[string(b)] = model.SocketLogin{
		Status: 1,
	}
	return encrypted
}

func GenerateImageQrcodeHandler(c *gin.Context) {
	var png []byte
	code := c.Param("code")

	c.Writer.Header().Set("Content-Type", "image/png")
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Printf("Length is %d bytes long\n", len(png))
	}
	c.Writer.Write(png)
}

func StartSocketHandler(c *gin.Context) {
	code := c.Param("code")
	decrypted, err := util.Decrypt(code)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	socket, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error while upgrading socket %v", err.Error())
		return
	}

	if tokens[decrypted].Status != 1 {
		defer socket.Close()
		err := socket.WriteMessage(websocket.TextMessage, []byte("Token is not active"))
		if err != nil {
			return
		}
	}

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
