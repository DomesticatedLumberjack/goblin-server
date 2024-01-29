package main

import (
	"github.com/gorilla/websocket"
)

func main() {
	server := Server{
		Upgrader: websocket.Upgrader{},
	}

	err := server.Run()
	if err != nil {
		panic(err)
	}
}
