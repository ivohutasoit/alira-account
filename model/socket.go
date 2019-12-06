package model

import "github.com/gorilla/websocket"

type LoginSocket struct {
	Callback string
	Status   int
	Socket   *websocket.Conn
}

var Sockets = make(map[string]LoginSocket)
