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
	Rooms map[string]*Room[DisasterMasterData]
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*Room[DisasterMasterData]),
	}
}

func (g *RoomManager) NewRoom(hostConn *websocket.Conn, hostId uint64, roomData DisasterMasterData) *Room[DisasterMasterData] {
	code := genCode()
	g.Rooms[code] = &Room[DisasterMasterData]{
		Code:     code,
		HostId:   hostId,
		HostConn: hostConn,

		RoomData: roomData,
	}

	return g.Rooms[code]
}

func (g *RoomManager) RemoveRoom(code string) {
	delete(g.Rooms, code)
}

func (g *RoomManager) JoinRoom(code string, conn *websocket.Conn, id uint64, goblin Goblin) error {
	// Check if the room exists
	_, ok := g.Rooms[code]
	if !ok {
		return ErrRoomNotFound{}
	}

	// Check if the client is already in the room
	for _, client := range g.Rooms[code].Clients {
		if client.Id == id {
			return nil
		}
	}

	// Remove the client from any existing Rooms
	g.RemoveClient(id)

	client := Client[Goblin]{
		Conn:       conn,
		Id:         id,
		PlayerData: goblin,
	}

	g.Rooms[code].Clients = append(g.Rooms[code].Clients, client)
	return nil
}

func (g *RoomManager) RemoveClient(id uint64) {
	for _, room := range g.Rooms {
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
	room, ok := g.Rooms[code]
	if !ok {
		return nil, ErrRoomNotFound{}
	}

	return room, nil
}
