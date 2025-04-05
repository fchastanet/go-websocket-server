package websocket

import (
	"encoding/json"
	"log"

	"learnLoop/main/models"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Clients organized by session ID
	sessionClients map[string][]*Client

	sessions map[string]*models.Session

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		sessions:       make(map[string]*models.Session),
		sessionClients: make(map[string][]*Client),
	}
}

func (hub *Hub) GetUsersInSession(sessionID string) []*models.User {
	clients := hub.sessionClients[sessionID]
	users := make([]*models.User, len(clients))
	for i, c := range clients {
		users[i] = c.User
	}
	return users
}

func SendMessageToAllClients(hub *Hub, msg any) error {
	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling JsonMessage: %v\n", err)
		return err
	}
	hub.broadcast <- jsonMessage
	return nil
}

func (h *Hub) GetSession(sessionId string) *models.Session {
	return h.sessions[sessionId]
}

func removeClient(h *Hub, client *Client) {
	log.Printf("unregistering client for user %s in session %s", client.User.UserID, client.User.SessionID)
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Remove client from session
		if clients, ok := h.sessionClients[client.User.SessionID]; ok {
			for i, c := range clients {
				if c == client {
					h.sessionClients[client.User.SessionID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
		}
	}
}

func (h *Hub) Run() {
	log.Println("hub is running")
	for {
		select {
		case client := <-h.register:
			log.Printf("registering client for user %s in session %s", client.User.UserID, client.User.SessionID)
			h.clients[client] = true

			// Associate client with session ID
			h.sessionClients[client.User.SessionID] = append(
				h.sessionClients[client.User.SessionID],
				client,
			)

			h.sessions[client.User.SessionID] = &models.Session{
				SessionID: client.User.SessionID,
				QuizGame: &models.QuizGame{
					SessionID: client.User.SessionID,
					GetConnectedPlayersCount: func() int {
						return len(h.sessionClients[client.User.SessionID])
					},
				},
			}

		case client := <-h.unregister:
			removeClient(h, client)
		case message := <-h.broadcast:
			log.Printf("sending message '%s' to all clients(%d)", message, len(h.clients))
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					removeClient(h, client)
				}
			}
		}
	}
}
