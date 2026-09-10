package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/Fedoroff05/auto-backend/internal/usecase"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSMessage struct {
	ChatID      uuid.UUID `json:"chat_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	UserID      uuid.UUID
	ChatUsecase *usecase.ChatUsecase
	send        chan *WSMessage
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming WSMessage
		if err := json.Unmarshal(payload, &incoming); err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		savedMsg, err := c.ChatUsecase.SaveNewMessage(ctx, incoming.ChatID, c.UserID, incoming.Content)
		cancel()
		if err != nil {
			log.Printf("Failed to save message: %v", err)
			continue
		}

		incoming.SenderID = c.UserID
		incoming.CreatedAt = savedMsg.CreatedAt
		c.Hub.broadcast <- &incoming
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// производит апгрейд HTTP до WebSocket
func ServeWS(hub *Hub, chatUsecase *usecase.ChatUsecase, userID uuid.UUID, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}

	client := &Client{
		Hub:         hub,
		Conn:        conn,
		UserID:      userID,
		ChatUsecase: chatUsecase,
		send:        make(chan *WSMessage, 256),
	}

	client.Hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
