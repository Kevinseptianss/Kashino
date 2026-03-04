package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"kashino-backend/internal/models"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096 // Increased from 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096, // Increased
	WriteBufferSize: 4096, // Increased
	CheckOrigin: func(r *http.Request) bool {
		return true // For development
	},
}

type Client struct {
	Hub      *Hub
	ID       string
	UserID   primitive.ObjectID
	Username string
	Conn     *websocket.Conn
	send     chan []byte
}

type WSMessage struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type WSResponse struct {
	Action string      `json:"action"`
	Status string      `json:"status"` // "success" or "error"
	Data   interface{} `json:"data"`
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshalling ws message: %v", err)
			continue
		}

		c.handleAction(msg)
	}
}

func (c *Client) handleAction(msg WSMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("WS Action received: %s from %s", msg.Action, c.ID)

	switch msg.Action {
	case "signup":
		var user models.User
		if err := json.Unmarshal(msg.Data, &user); err != nil {
			c.sendError("signup", "Invalid user data")
			return
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)
		if err := c.Hub.UserRepo.Create(ctx, &user); err != nil {
			c.sendError("signup", "Failed to create user")
			return
		}
		c.sendSuccess("signup", user)

	case "login":
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(msg.Data, &creds); err != nil {
			c.sendError("login", "Invalid credentials data")
			return
		}
		user, err := c.Hub.UserRepo.FindByUsername(ctx, creds.Username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)) != nil {
			c.sendError("login", "Invalid username or password")
			return
		}
		c.ID = user.Username
		c.UserID = user.ID
		c.sendSuccess("login", map[string]interface{}{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"balance":  user.Balance,
			"tier":     user.Tier,
		})

	case "get_balance":
		if c.UserID.IsZero() {
			c.sendError("get_balance", "Unauthorized")
			return
		}
		user, err := c.Hub.UserRepo.GetUser(ctx, c.UserID)
		if err != nil {
			c.sendError("get_balance", "User not found")
			return
		}
		c.sendSuccess("get_balance", map[string]interface{}{
			"balance": user.Balance,
			"tier":    user.Tier,
		})

	case "get_history":
		if c.UserID.IsZero() {
			c.sendError("get_history", "Unauthorized")
			return
		}
		user, err := c.Hub.UserRepo.GetUser(ctx, c.UserID)
		if err != nil {
			c.sendError("get_history", "User not found")
			return
		}
		c.sendSuccess("get_history", map[string]interface{}{
			"history": user.BalanceHistory,
		})

	case "update_balance":
		if c.UserID.IsZero() {
			c.sendError("update_balance", "Unauthorized")
			return
		}
		var data struct {
			Amount int64  `json:"amount"`
			Source string `json:"source"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError("update_balance", "Invalid data")
			return
		}
		err := c.Hub.UserRepo.UpdateBalance(ctx, c.UserID, data.Amount, data.Source)
		if err != nil {
			c.sendError("update_balance", "Failed to update balance")
			return
		}
		// Fetch new balance to return
		newBalance, _ := c.Hub.UserRepo.GetBalance(ctx, c.UserID)
		c.sendSuccess("balance_update", map[string]interface{}{
			"balance": newBalance,
		})

	case "get_rooms":
		rooms := c.Hub.GetRoomList()
		log.Printf("Sending %d rooms to client %s", len(rooms), c.ID)
		c.sendSuccess("get_rooms", rooms)

	case "join_room":
		var joinData struct {
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(msg.Data, &joinData); err != nil {
			c.sendError("join_room", "Invalid join data")
			return
		}
		room, ok := c.Hub.GetRoom(joinData.RoomID)
		if !ok {
			c.sendError("join_room", "Room not found")
			return
		}
		c.sendSuccess("join_room", room)
		c.Hub.SendChatHistory(c, joinData.RoomID)

	case "chat_message":
		if c.UserID.IsZero() {
			c.sendError("chat_message", "Unauthorized")
			return
		}
		var chatData struct {
			RoomID  string `json:"room_id"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(msg.Data, &chatData); err != nil {
			c.sendError("chat_message", "Invalid chat data")
			return
		}

		chatMsg := models.ChatMessage{
			ID:        primitive.NewObjectID(),
			RoomID:    chatData.RoomID,
			UserID:    c.UserID.Hex(),
			Username:  c.Username,
			Message:   chatData.Message,
			Timestamp: primitive.NewDateTimeFromTime(time.Now()),
		}

		c.Hub.HandleChatMessage(chatMsg)

	case "public_chat_message":
		if c.UserID.IsZero() {
			c.sendError("public_chat_message", "Unauthorized")
			return
		}
		var chatData struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(msg.Data, &chatData); err != nil {
			c.sendError("public_chat_message", "Invalid chat data")
			return
		}

		chatMsg := models.ChatMessage{
			ID:        primitive.NewObjectID(),
			RoomID:    "public",
			UserID:    c.UserID.Hex(),
			Username:  c.Username,
			Message:   chatData.Message,
			Timestamp: primitive.NewDateTimeFromTime(time.Now()),
		}

		c.Hub.HandleChatMessage(chatMsg)

	case "public_chat_history":
		c.Hub.SendChatHistory(c, "public")

	case "sit":
		var sitData struct {
			RoomID string `json:"room_id"`
			Seat   int    `json:"seat"`
		}
		if err := json.Unmarshal(msg.Data, &sitData); err != nil {
			c.sendError("sit", "Invalid sit data")
			return
		}

		// Check if room exists first
		_, ok := c.Hub.GetRoom(sitData.RoomID)
		if !ok {
			c.sendError("sit", "Room not found")
			return
		}

		// Unmarshal full user to get balance
		user, _ := c.Hub.UserRepo.GetUser(ctx, c.UserID)

		player := models.Player{
			ID:       c.UserID.Hex(),
			Username: user.Username,
			Balance:  user.Balance,
			Position: sitData.Seat,
			InGame:   true,
		}

		c.Hub.SitPlayer(sitData.RoomID, sitData.Seat, player)

	case "poker_action":
		var actionData struct {
			RoomID string             `json:"room_id"`
			Action models.PokerAction `json:"action"`
		}
		if err := json.Unmarshal(msg.Data, &actionData); err != nil {
			c.sendError("poker_action", "Invalid action data")
			return
		}

		c.Hub.HandlePokerAction(actionData.RoomID, c.UserID.Hex(), actionData.Action)

	case "standup":
		c.Hub.RemovePlayerFromAllRooms(c.UserID.Hex())
		c.sendSuccess("standup", nil)

	case "slot_log":
		var slotData struct {
			Bet       int64   `json:"bet"`
			Lines     int     `json:"lines"`
			Result    [][]int `json:"result"`
			Winners   [][]int `json:"winners"`
			WinAmount int64   `json:"win_amount"`
		}
		if err := json.Unmarshal(msg.Data, &slotData); err != nil {
			c.sendError("slot_log", "Invalid slot log data")
			return
		}

		// Fetch username for logging
		user, _ := c.Hub.UserRepo.GetUser(ctx, c.UserID)
		username := "Unknown"
		if user != nil {
			username = user.Username
		}

		c.Hub.LogSlotEvent(c.UserID.Hex(), username, slotData.Bet, slotData.Lines, slotData.Result, slotData.Winners, slotData.WinAmount)

		// Deduct bet and add win to balance
		if slotData.Bet > 0 {
			c.Hub.UpdateBalance(c.UserID.Hex(), -slotData.Bet, "slot_spin")
		}
		if slotData.WinAmount > 0 {
			c.Hub.UpdateBalance(c.UserID.Hex(), slotData.WinAmount, "slot_win")
		}

		c.sendSuccess("slot_log", slotData)

	default:
		log.Printf("Unknown action: %s", msg.Action)
	}
}

func (c *Client) sendSuccess(action string, data interface{}) {
	resp := WSResponse{
		Action: action,
		Status: "success",
		Data:   data,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling success response: %v", err)
		return
	}
	log.Printf("Sending WS Success: %s", string(b))
	c.send <- b
}

func (c *Client) sendError(action string, message string) {
	resp := WSResponse{
		Action: action,
		Status: "error",
		Data:   map[string]string{"message": message},
	}
	b, _ := json.Marshal(resp)
	c.send <- b
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			log.Printf("WS writePump: sending %d bytes to %s", len(message), c.ID)
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	var userID primitive.ObjectID
	var username string

	if userIDStr != "" {
		userID, _ = primitive.ObjectIDFromHex(userIDStr)
		if !userID.IsZero() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if user, err := hub.UserRepo.GetUser(ctx, userID); err == nil && user != nil {
				username = user.Username
			}
		}
	}

	if username == "" {
		if len(userIDStr) > 5 {
			username = "User_" + userIDStr[:5]
		} else {
			username = "Guest"
		}
	}

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		send:     make(chan []byte, 256),
		ID:       userIDStr,
		UserID:   userID,
		Username: username,
	}

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}
