package chat

import (
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	conn     *websocket.Conn
	username string
	send     chan roomState
}

var room = struct {
	sync.RWMutex
	clients map[*client]struct{}
}{clients: make(map[*client]struct{})}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(req *http.Request) bool { return true },
}

type incomingMessage struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Message string `json:"message"`
}

func HandleWebSocket(w http.ResponseWriter, req *http.Request) {
	sID, err := req.Cookie("chat_sid")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionMu.RLock()
	username := Sessions[sID.Value]
	sessionMu.RUnlock()
	if username == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}

	c := &client{conn: conn, username: username, send: make(chan roomState, 16)}
	room.Lock()
	room.clients[c] = struct{}{}
	room.Unlock()

	go c.writePump()
	c.send <- roomState{Type: "state", Messages: currentMessages(username), Users: connectedUsers()}
	broadcastUsers()
	c.readPump()
}

func (c *client) readPump() {
	defer func() {
		room.Lock()
		delete(room.clients, c)
		room.Unlock()
		c.conn.Close()
		broadcastUsers()
	}()

	for {
		var input incomingMessage
		if err := c.conn.ReadJSON(&input); err != nil {
			return
		}
		if input.Type != "message" || input.Message == "" || input.To == "" {
			continue
		}

		fromUser := GetUser(c.username)
		toUser := GetUser(input.To)
		if toUser.ID == 0 || SaveMessage(fromUser, toUser, input.Message) != nil {
			continue
		}

		message := TextMessage{From: c.username, To: input.To, Message: input.Message}
		broadcast(roomState{Type: "message", Message: &message})
	}
}

func (c *client) writePump() {
	defer c.conn.Close()
	for state := range c.send {
		if err := c.conn.WriteJSON(state); err != nil {
			return
		}
	}
}

func broadcast(state roomState) {
	room.RLock()
	defer room.RUnlock()
	for c := range room.clients {
		select {
		case c.send <- state:
		default:
			log.Printf("dropping websocket event for %s", c.username)
		}
	}
}

func broadcastUsers() {
	broadcast(roomState{Type: "users", Users: connectedUsers()})
}

func connectedUsers() []string {
	room.RLock()
	defer room.RUnlock()
	users := make([]string, 0, len(room.clients))
	seen := make(map[string]struct{})
	for c := range room.clients {
		if _, exists := seen[c.username]; exists {
			continue
		}
		seen[c.username] = struct{}{}
		users = append(users, c.username)
	}
	sort.Strings(users)
	return users
}

func currentMessages(username string) []TextMessage {
	user := GetUser(username)
	return messageIDsToText(GetMessages(user, user))
}
