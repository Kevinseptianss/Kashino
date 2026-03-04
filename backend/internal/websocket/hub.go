package websocket

import (
	"context"
	"encoding/json"
	"kashino-backend/internal/models"
	"kashino-backend/internal/poker"
	"kashino-backend/internal/repository"
	"log"
	"time"

	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Hub struct {
	UserRepo       *repository.UserRepository
	PokerRepo      *repository.PokerRepository
	ChatRepo       *repository.ChatRepository
	clients        map[*Client]bool
	Rooms          map[string]*models.Room
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	balanceUpdates chan balanceUpdateMsg
	mu             sync.RWMutex
}

type balanceUpdateMsg struct {
	UserID  primitive.ObjectID
	Balance int64
}

func NewHub(userRepo *repository.UserRepository, pokerRepo *repository.PokerRepository, chatRepo *repository.ChatRepository) *Hub {
	h := &Hub{
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		Rooms:          make(map[string]*models.Room),
		balanceUpdates: make(chan balanceUpdateMsg, 1024),
		UserRepo:       userRepo,
		PokerRepo:      pokerRepo,
		ChatRepo:       chatRepo,
	}

	// Create default rooms
	h.Rooms["room1"] = &models.Room{
		ID:         "room1",
		Name:       "Table 1",
		MaxPlayers: 5,
		SmallBlind: 10,
		BigBlind:   20,
		GameState: models.GameState{
			ID:      "room1",
			Players: make([]models.Player, 0),
			Round:   "waiting",
		},
	}
	h.Rooms["room2"] = &models.Room{
		ID:         "room2",
		Name:       "Table 2",
		MaxPlayers: 5,
		SmallBlind: 50,
		BigBlind:   100,
		GameState: models.GameState{
			ID:      "room2",
			Players: make([]models.Player, 0),
			Round:   "waiting",
		},
	}
	h.Rooms["room3"] = &models.Room{
		ID:         "room3",
		Name:       "Table 3",
		MaxPlayers: 5,
		SmallBlind: 100,
		BigBlind:   200,
		GameState: models.GameState{
			ID:      "room3",
			Players: make([]models.Player, 0),
			Round:   "waiting",
		},
	}
	log.Printf("Initialized Hub with %d rooms", len(h.Rooms))

	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Disconnect any existing sessions for this UserID
			if !client.UserID.IsZero() {
				h.mu.Lock()
				for existing := range h.clients {
					if existing.UserID == client.UserID && existing != client {
						log.Printf("Disconnecting existing session for user %s (new connection)", client.ID)
						// Remove from rooms first
						h.removePlayerFromAllRoomsLocked(existing.UserID.Hex())
						// Then delete and close
						delete(h.clients, existing)
						close(existing.send)
					}
				}
				h.mu.Unlock()
			}
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.ID)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s", client.ID)

				// Remove player from all rooms
				h.removePlayerFromAllRoomsLocked(client.UserID.Hex())
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Instead of immediate delete/close in RLock,
					// we should ideally use a channel to unregister
					log.Printf("Warning: Client %s buffer full, message dropped", client.ID)
				}
			}
			h.mu.RUnlock()
		case upd := <-h.balanceUpdates:
			h.mu.RLock()
			resp := WSResponse{
				Action: "balance_update",
				Status: "success",
				Data:   map[string]interface{}{"balance": upd.Balance},
			}
			data, _ := json.Marshal(resp)
			for client := range h.clients {
				if client.UserID == upd.UserID {
					select {
					case client.send <- data:
					default:
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshalling broadcast JSON: %v", err)
		return
	}
	h.broadcast <- data
}

func (h *Hub) GetRoomList() []models.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rooms := make([]models.Room, 0, len(h.Rooms))
	for _, room := range h.Rooms {
		rooms = append(rooms, *room)
	}
	return rooms
}

func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) GetRoom(id string) (*models.Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, ok := h.Rooms[id]
	return room, ok
}

func (h *Hub) SitPlayer(roomID string, seat int, player models.Player) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.Rooms[roomID]
	if !ok {
		return
	}

	// Remove player from any other seat in any room
	h.removePlayerFromAllRoomsLocked(player.ID)

	// Check if seat is taken
	for _, p := range room.GameState.Players {
		if p.Position == seat {
			return
		}
	}

	room.GameState.Players = append(room.GameState.Players, player)
	poker.StartHand(room, h)

	h.broadcastRoomUpdateLocked("player_sat", room)
}

func (h *Hub) HandlePokerAction(roomID string, userID string, action models.PokerAction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.Rooms[roomID]
	if !ok {
		return
	}

	poker.HandleAction(room, userID, action, h)
	h.broadcastRoomUpdateLocked("room_update", room)

	// If the round just ended, schedule the next hand
	if room.GameState.Round == "waiting" {
		// Use a local variable to capture roomID and handID to avoid race later
		hid := room.GameState.ID
		go h.scheduleNextHand(roomID, hid)
	}
}

func (h *Hub) scheduleNextHand(roomID string, triggerHandID string) {
	// First sleep a bit to allow final updates to settle
	time.Sleep(500 * time.Millisecond)

	h.mu.Lock()
	room, ok := h.Rooms[roomID]
	// If room changed round already or hand ID changed, stop
	if !ok || room.GameState.Round != "waiting" || room.GameState.ID != triggerHandID {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	for i := 10; i > 0; i-- {
		h.mu.Lock()
		room, ok = h.Rooms[roomID]
		// Double check we are still in the same waiting state
		if !ok || room.GameState.Round != "waiting" || room.GameState.ID != triggerHandID {
			h.mu.Unlock()
			return
		}
		room.GameState.NextHandCountdown = i
		h.broadcastRoomUpdateLocked("room_update", room)
		h.mu.Unlock()
		time.Sleep(1 * time.Second)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	room, ok = h.Rooms[roomID]
	if ok && room.GameState.Round == "waiting" && room.GameState.ID == triggerHandID {
		room.GameState.NextHandCountdown = 0
		poker.StartHand(room, h)
		h.broadcastRoomUpdateLocked("room_update", room)
	}
}

func (h *Hub) RemovePlayerFromAllRooms(userIDHex string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removePlayerFromAllRoomsLocked(userIDHex)
}

func (h *Hub) NotifyRoomUpdate(action string, room *models.Room) {
	h.BroadcastRoomUpdate(action, room)
}

func (h *Hub) BroadcastRoomUpdate(action string, room *models.Room) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.broadcastRoomUpdateLocked(action, room)
}

func (h *Hub) broadcastRoomUpdateLocked(action string, room *models.Room) {
	for client := range h.clients {
		// Create a copy of the room state for sanitization
		sanitizedRoom := *room
		sanitizedGameState := room.GameState
		sanitizedPlayers := make([]models.Player, len(room.GameState.Players))

		for i, p := range room.GameState.Players {
			sanitizedPlayer := p
			// Hide cards and hand ranking of OTHER players
			// Unless it's showdown or the player is themselves
			userID := client.UserID.Hex()
			if p.ID != userID && room.GameState.Round != "showdown" && room.GameState.Round != "waiting" {
				sanitizedPlayer.Cards = nil
				sanitizedPlayer.CurrentHand = ""
			}
			sanitizedPlayers[i] = sanitizedPlayer
		}

		sanitizedGameState.Players = sanitizedPlayers
		sanitizedRoom.GameState = sanitizedGameState

		resp := WSResponse{
			Action: action,
			Status: "success",
			Data:   sanitizedRoom,
		}
		data, _ := json.Marshal(resp)

		select {
		case client.send <- data:
		default:
		}
	}
}

func (h *Hub) removePlayerFromAllRoomsLocked(userIDHex string) {
	for _, room := range h.Rooms {
		poker.LeaveGame(room, userIDHex, h)
		h.broadcastRoomUpdateLocked("room_update", room)
	}
}

func (h *Hub) UpdateBalance(userID string, amount int64, source string) {
	objID, _ := primitive.ObjectIDFromHex(userID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.UserRepo.UpdateBalance(ctx, objID, amount, source)
	if err != nil {
		log.Printf("Error updating balance for user %s: %v", userID, err)
		return
	}

	// Fetch new balance and notify connected clients asynchronously
	newBalance, err := h.UserRepo.GetBalance(ctx, objID)
	if err == nil {
		select {
		case h.balanceUpdates <- balanceUpdateMsg{UserID: objID, Balance: newBalance}:
		default:
			log.Printf("Warning: balanceUpdates channel full, notification dropped for %s", userID)
		}
	}
}

func (h *Hub) LogPokerEvent(roomID string, handID string, event string, playerID string, username string, amount int64, pot int64, cards []models.Card, community []models.Card, details string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if handID == "" {
		handID = "h_" + primitive.NewObjectID().Hex()
	}

	history := models.PokerHistory{
		ID:        primitive.NewObjectID(),
		RoomID:    roomID,
		HandID:    handID,
		Event:     event,
		PlayerID:  playerID,
		Username:  username,
		Amount:    amount,
		Pot:       pot,
		Cards:     cards,
		Community: community,
		Details:   details,
		Timestamp: primitive.NewDateTimeFromTime(time.Now()),
	}

	h.PokerRepo.LogEvent(ctx, history)
}

func (h *Hub) HandleChatMessage(msg models.ChatMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Save to DB
	err := h.ChatRepo.SaveMessage(ctx, msg)
	if err != nil {
		log.Printf("Error saving chat message: %v", err)
	}

	// Broadcast to all clients
	action := "chat_message"
	if msg.RoomID == "public" {
		action = "public_chat_message"
	}

	resp := WSResponse{
		Action: action,
		Status: "success",
		Data:   msg,
	}
	h.BroadcastJSON(resp)
}

func (h *Hub) SendChatHistory(client *Client, roomID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	history, err := h.ChatRepo.GetHistory(ctx, roomID, 50)
	if err != nil {
		log.Printf("Error getting chat history: %v", err)
		return
	}

	action := "chat_history"
	if roomID == "public" {
		action = "public_chat_history"
	}

	resp := WSResponse{
		Action: action,
		Status: "success",
		Data:   map[string]interface{}{"room_id": roomID, "messages": history},
	}
	data, _ := json.Marshal(resp)
	client.send <- data
}

func (h *Hub) LogSlotEvent(userID string, username string, bet int64, lines int, result [][]int, winners [][]int, win int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	history := models.SlotHistory{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Username:  username,
		Bet:       bet,
		Lines:     lines,
		Result:    result,
		Winners:   winners,
		WinAmount: win,
		Timestamp: primitive.NewDateTimeFromTime(time.Now()),
	}

	h.PokerRepo.LogSlotEvent(ctx, history)
}
