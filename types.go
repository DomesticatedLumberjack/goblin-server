package main

import "github.com/gorilla/websocket"

// General types
type Room[T any] struct {
	Code     string          `json:"code"`
	HostId   uint64          `json:"hostId"`
	HostConn *websocket.Conn `json:"-"`

	RoomData T `json:"disasterMasterData"`

	Clients []Client[Goblin] `json:"clients"`
}

type Client[T any] struct {
	Conn       *websocket.Conn `json:"-"`
	Id         uint64          `json:"id"`
	PlayerData T               `json:"playerData"`
}

type Message struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}

// Game specific data definition
type DisasterMasterData struct {
	ScenarioCode string `json:"scenario"`
	ChaosClock   int    `json:"chaosClock"`
}

type Goblin struct {
	Dice           []int  `json:"dice"`
	Name           string `json:"name"`
	AssSize        int    `json:"assSize"`
	AssOrigin      int    `json:"assOrigin"`
	Class          int    `json:"class"`
	PocketContents string `json:"pocketContents"`
	Note           string `json:"note"`
}

// Errors
type ErrRoomNotFound struct{}

func (e ErrRoomNotFound) Error() string {
	return "Room not found"
}

type ErrRoomAlreadyExists struct{}

func (e ErrRoomAlreadyExists) Error() string {
	return "Room already exists"
}

type ErrRoomFull struct{}

func (e ErrRoomFull) Error() string {
	return "Room is full"
}

type ErrInvalidCommand struct{}

func (e ErrInvalidCommand) Error() string {
	return "Invalid command"
}

type ErrUnableToWrite struct{}

func (e ErrUnableToWrite) Error() string {
	return "Unable to write"
}

type ErrUnableToRead struct{}

func (e ErrUnableToRead) Error() string {
	return "Unable to read"
}

type ErrMissingField struct {
	Field string
}

func (e ErrMissingField) Error() string {
	return "Missing field: " + e.Field
}
