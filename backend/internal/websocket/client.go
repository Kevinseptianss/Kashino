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
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development
	},
}

type Client struct {
	Hub    *Hub
	ID     string
	UserID primitive.ObjectID
	Conn   *websocket.Conn
	send   chan []byte
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
	b, _ := json.Marshal(resp)
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
	if userIDStr != "" {
		userID, _ = primitive.ObjectIDFromHex(userIDStr)
	}

	client := &Client{
		Hub:    hub,
		Conn:   conn,
		send:   make(chan []byte, 256),
		ID:     userIDStr,
		UserID: userID,
	}
	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}
