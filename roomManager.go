package main

import (
	"math/rand"

	"github.com/gorilla/websocket"
)

var genChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func genCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		code += string(genChars[rand.Intn(len(genChars))])
	}

	return code
}

type RoomManager struct {
	rooms map[string]*Room[DisasterMasterData]
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room[DisasterMasterData]),
	}
}

func (g *RoomManager) NewRoom(hostConn *websocket.Conn, hostId uint64, roomData DisasterMasterData) *Room[DisasterMasterData] {
	code := genCode()
	g.rooms[code] = &Room[DisasterMasterData]{
		Code:     code,
		HostId:   hostId,
		HostConn: hostConn,

		RoomData: roomData,
	}

	return g.rooms[code]
}

func (g *RoomManager) RemoveRoom(code string) {
	delete(g.rooms, code)
}

func (g *RoomManager) JoinRoom(code string, conn *websocket.Conn, id uint64, goblin Goblin) error {
	// Check if the room exists
	_, ok := g.rooms[code]
	if !ok {
		return ErrRoomNotFound{}
	}

	// Check if the client is already in the room
	for _, client := range g.rooms[code].Clients {
		if client.Id == id {
			return nil
		}
	}

	// Remove the client from any existing rooms
	g.RemoveClient(id)

	client := Client[Goblin]{
		Conn:       conn,
		Id:         id,
		PlayerData: goblin,
	}

	g.rooms[code].Clients = append(g.rooms[code].Clients, client)
	return nil
}

func (g *RoomManager) RemoveClient(id uint64) {
	for _, room := range g.rooms {
		for i, client := range room.Clients {
			if client.Id == id {
				room.Clients = append(room.Clients[:i], room.Clients[i+1:]...)

				return
			}
		}
	}
}

// Lol
func (g *RoomManager) GetRoom(code string) (*Room[DisasterMasterData], error) {
	room, ok := g.rooms[code]
	if !ok {
		return nil, ErrRoomNotFound{}
	}

	return room, nil
}
