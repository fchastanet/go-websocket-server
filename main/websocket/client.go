package websocket

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"learnLoop/main/auth"
	"learnLoop/main/models"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type (
	CloseHandler   func(client *Client) error
	MessageHandler func(client *Client, message []byte) error
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	User *models.User

	CloseHandler   CloseHandler
	MessageHandler MessageHandler
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.conn.Close()
		err := c.CloseHandler(c)
		if err != nil {
			log.Printf("Error sending UserDisconnectMessage: %v\n", err)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()

		if _, ok := err.(*websocket.CloseError); ok {
			log.Printf("socket closed: %v", err)
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		err = c.MessageHandler(c, message)
		if err != nil {
			log.Printf("Error handling message: %v\n", err)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write the message to the WebSocket connection
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(
	hub *Hub, w http.ResponseWriter, r *http.Request,
	clientCloseHandler CloseHandler, clientMessageHandler MessageHandler,
) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Extract JWT token from query parameters
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	// Validate JWT token
	userID, instanceName, err := auth.ValidateJWT(token)
	if err != nil {
		log.Printf("Invalid JWT token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Authenticated user %s for session %s on instance %s", userID, sessionID, instanceName)

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		Hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		User: &models.User{
			Login:        "",
			UserID:       userID,
			SessionID:    sessionID,
			InstanceName: instanceName,
		},
		CloseHandler:   clientCloseHandler,
		MessageHandler: clientMessageHandler,
	}

	// Register client with session ID
	client.Hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
