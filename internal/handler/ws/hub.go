package ws

import (
	"sync"

	"github.com/google/uuid"
)

// управляет всеми активными соединениями клиентов
type Hub struct {
	//cопоставление userID -> список открытых клиентских сокетов
	clients    map[uuid.UUID]map[*Client]bool
	broadcast  chan *WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan *WSMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// крутится в отдельной горутине на протяжении всей жизни сервера
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if userClients, ok := h.clients[client.UserID]; ok {
				if _, exists := userClients[client]; exists {
					delete(userClients, client)
					close(client.send)
					if len(userClients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			//отправляем получателю
			if userClients, ok := h.clients[message.RecipientID]; ok {
				for client := range userClients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(userClients, client)
					}
				}
			}
			//отправляем обратно отправителю
			if userClients, ok := h.clients[message.SenderID]; ok {
				for client := range userClients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(userClients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}
