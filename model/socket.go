package model

import "github.com/gorilla/websocket"

type SocketLogin struct {
	Status int
	Socket *websocket.Conn
}