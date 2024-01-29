package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const maxUint64 = ^uint64(0) //Change this to guid later

type Server struct {
	upgrader          websocket.Upgrader
	connectionCounter uint64
	RoomManager       *RoomManager
}

func (s *Server) Run() error {
	s.RoomManager = NewRoomManager()
	s.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	//Socket
	http.HandleFunc("/ws", s.handleSocket)

	//Test Html page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test.html")
	})

	log.Println("Goblin server running on port 8080...")
	log.Println("Connect on /")

	return http.ListenAndServe(":8080", nil)
}

func (s *Server) handleSocket(w http.ResponseWriter, r *http.Request) {
	//Upgrade to websocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade failed: ", err)
		return
	}
	defer conn.Close()

	//Handle assigning ids
	connId := s.connectionCounter
	if s.connectionCounter == maxUint64 {
		s.connectionCounter = 0
	} else {
		s.connectionCounter++
	}
	log.Println("Connection opened by:", connId)

	//Listen for messages
	for {
		log.Println("Waiting for message from:", connId)
		// Read the message
		var body Message
		err := conn.ReadJSON(&body)
		if err != nil {
			log.Println("Connection closed by", connId)
			break
		}

		log.Println("Handling", body.Command, "from", connId)

		// Handle the message
		switch body.Command {
		case "create":
			s.create(conn, connId, body)
		case "join":
			s.join(conn, connId, body)
		case "updateplayer":
			s.updatePlayer(conn, connId, body)
		case "updateroom":
			s.updateRoom(conn, connId, body)
		default:
			s.handleError(conn, connId, ErrInvalidCommand{})
		}

		log.Println("Current Games:", s.RoomManager.rooms)
	}
}

func (s *Server) create(conn *websocket.Conn, id uint64, body Message) {
	log.Println("Create command received from:", id)

	scenarioCode, ok := body.Data["scenarioCode"].(string)
	if !ok {
		s.handleError(conn, id, ErrMissingField{"scenarioCode"})
		return
	}

	// Create a new game
	room := s.RoomManager.NewRoom(conn, id, DisasterMasterData{
		ScenarioCode: scenarioCode,
		ChaosClock:   0,
	})

	// Delete the game when the connection closes
	conn.SetCloseHandler(func(code int, text string) error {
		s.RoomManager.RemoveRoom(room.Code)

		// Send a response to all clients in the game
		for _, client := range room.Clients {
			err := client.Conn.WriteJSON(Message{
				Command: "quit",
				Data: map[string]interface{}{
					"message": "Game closed",
				},
			})
			if err != nil {
				s.handleError(client.Conn, client.Id, ErrUnableToWrite{})
			}
		}

		return nil
	})

	s.sendGameState(room)
}

func (s *Server) join(conn *websocket.Conn, id uint64, body Message) {
	log.Println("Join command received from:", id)

	code, ok := body.Data["code"].(string)
	if !ok {
		s.handleError(conn, id, ErrMissingField{"code"})
		return
	}

	playerData, ok := body.Data["player"].(map[string]interface{})
	if !ok {
		s.handleError(conn, id, ErrMissingField{"player"})
		return
	}

	//Handle dice
	dice, ok := playerData["dice"].([]interface{})
	if !ok {
		s.handleError(conn, id, ErrMissingField{"dice"})
		return
	}
	diceInt := make([]int, len(dice))
	for i, v := range dice {
		diceInt[i] = int(v.(float64))
	}

	//Convert map to Goblin
	goblin := Goblin{
		Dice:           diceInt,
		Name:           playerData["name"].(string),
		AssSize:        int(playerData["assSize"].(float64)),
		AssOrigin:      int(playerData["assOrigin"].(float64)),
		Class:          int(playerData["class"].(float64)),
		PocketContents: playerData["pocketContents"].(string),
		Note:           playerData["note"].(string),
	}

	// Get the room
	room, err := s.RoomManager.GetRoom(code)
	if err != nil {
		s.handleError(conn, id, err)
		return
	}

	// Add the client to the room
	s.RoomManager.JoinRoom(room.Code, conn, id, goblin)

	// Remove the client from the room when the connection closes
	conn.SetCloseHandler(func(code int, text string) error {
		s.RoomManager.RemoveClient(id)
		s.sendGameState(room)
		return nil
	})

	s.sendGameState(room)
}

func (s *Server) updateRoom(conn *websocket.Conn, id uint64, body Message) {
	log.Println("Update command received from:", id)

	code, ok := body.Data["code"].(string)
	if !ok {
		s.handleError(conn, id, ErrMissingField{"code"})
		return
	}

	chaosClock, ok := body.Data["chaosClock"].(float64)
	if !ok {
		s.handleError(conn, id, ErrMissingField{"chaosClock"})
		return
	}

	scenarioCode, ok := body.Data["scenarioCode"].(string)
	if !ok {
		s.handleError(conn, id, ErrMissingField{"scenarioCode"})
		return
	}

	room, err := s.RoomManager.GetRoom(code)
	if err != nil {
		s.handleError(conn, id, err)
		return
	}

	room.RoomData.ChaosClock = int(chaosClock)
	room.RoomData.ScenarioCode = scenarioCode

	s.sendGameState(room)
}

func (s *Server) updatePlayer(conn *websocket.Conn, id uint64, body Message) {
	code, ok := body.Data["code"].(string)
	if !ok {
		s.handleError(conn, id, ErrMissingField{"code"})
		return
	}

	playerData, ok := body.Data["player"].(map[string]interface{})
	if !ok {
		s.handleError(conn, id, ErrMissingField{"player"})
		return
	}

	//Handle dice
	dice, ok := playerData["dice"].([]interface{})
	if !ok {
		s.handleError(conn, id, ErrMissingField{"dice"})
		return
	}
	diceInt := make([]int, len(dice))
	for i, v := range dice {
		diceInt[i] = int(v.(float64))
	}

	//Convert map to Goblin
	goblin := Goblin{
		Dice:           diceInt,
		Name:           playerData["name"].(string),
		AssSize:        int(playerData["assSize"].(float64)),
		AssOrigin:      int(playerData["assOrigin"].(float64)),
		Class:          int(playerData["class"].(float64)),
		PocketContents: playerData["pocketContents"].(string),
		Note:           playerData["note"].(string),
	}

	room, err := s.RoomManager.GetRoom(code)
	if err != nil {
		s.handleError(conn, id, err)
		return
	}

	for i, client := range room.Clients {
		if client.Id == id {
			room.Clients[i].PlayerData = goblin
			break
		}
	}

	s.sendGameState(room)
}

func (s *Server) sendGameState(room *Room[DisasterMasterData]) {
	for _, client := range room.Clients {

		// Filter out the target client
		filteredClients := make([]Client[Goblin], 0, len(room.Clients)-1)
		for _, c := range room.Clients {
			if c.Id != client.Id {
				filteredClients = append(filteredClients, c)
			}
		}

		err := client.Conn.WriteJSON(Message{
			Command: "update",
			Data: map[string]interface{}{
				"clientId": client.Id,
				"room": &Room[DisasterMasterData]{
					Code:     room.Code,
					HostId:   room.HostId,
					RoomData: room.RoomData,
					Clients:  filteredClients,
				},
			},
		})
		if err != nil {
			s.handleError(client.Conn, client.Id, ErrUnableToWrite{})
		}
	}

	err := room.HostConn.WriteJSON(Message{
		Command: "update",
		Data: map[string]interface{}{
			"clientId": room.HostId,
			"room":     room,
		},
	})
	if err != nil {
		s.handleError(room.HostConn, room.HostId, ErrUnableToWrite{})
	}
}

func (s *Server) handleError(conn *websocket.Conn, id uint64, err error) {
	log.Println("( ID:", id, ") Error:", err)

	// Send a response
	conn.WriteJSON(Message{
		Command: "error",
		Data: map[string]interface{}{
			"message": err.Error(),
		},
	})
}
