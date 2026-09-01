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
	c.send <- roomState{Type: "state", Contacts: connectedContacts()}
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
			log.Printf("websocket read for %s: %v", c.username, err)
			return
		}
		if input.Type != "message" || input.Message == "" || input.To == "" {
			if input.Type == "history" && input.To != "" {
				log.Printf("loading conversation for %s with %s", c.username, input.To)
				c.send <- roomState{Type: "history", ChatWith: input.To, Messages: conversationMessages(c.username, input.To)}
			}
			continue
		}

		fromUser := GetUser(c.username)
		toUser := GetUser(input.To)
		if fromUser.ID == 0 || toUser.ID == 0 {
			log.Printf("message rejected: %s -> %s, user not found", c.username, input.To)
			continue
		}
		if err := SaveMessage(fromUser, toUser, input.Message); err != nil {
			log.Printf("message rejected: %s -> %s: %v", c.username, input.To, err)
			continue
		}

		message := TextMessage{From: c.username, To: input.To, Message: input.Message}
		broadcastMessage(message)
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
	delivered := 0
	for c := range room.clients {
		select {
		case c.send <- state:
			delivered++
		default:
			log.Printf("dropping websocket event for %s", c.username)
		}
	}
	if state.Type == "message" {
		log.Printf("broadcast message to %d websocket clients", delivered)
	}
}

func broadcastUsers() {
	broadcast(roomState{Type: "contacts", Contacts: connectedContacts()})
}

func broadcastMessage(message TextMessage) {
	room.RLock()
	defer room.RUnlock()
	delivered := 0
	for c := range room.clients {
		if c.username != message.From && c.username != message.To {
			continue
		}
		messageForClient := message
		messageForClient.Mine = message.From == c.username
		state := roomState{Type: "message", Message: &messageForClient}
		select {
		case c.send <- state:
			delivered++
		default:
			log.Printf("dropping websocket message for %s", c.username)
		}
	}
	log.Printf("broadcast message to %d websocket clients", delivered)
}

func connectedContacts() []Contact {
	online := make(map[string]bool)
	room.RLock()
	for c := range room.clients {
		online[c.username] = true
	}
	room.RUnlock()

	users := GetUsers()
	contacts := make([]Contact, 0, len(users))
	for _, user := range users {
		contacts = append(contacts, Contact{Username: user.Username, Online: online[user.Username]})
	}
	sort.Slice(contacts, func(i, j int) bool { return contacts[i].Username < contacts[j].Username })
	return contacts
}

func conversationMessages(username, other string) []TextMessage {
	first := GetUser(username)
	second := GetUser(other)
	if first.ID == 0 || second.ID == 0 {
		return []TextMessage{}
	}
	messages := messageIDsToText(GetConversation(first, second))
	for i := range messages {
		messages[i].Mine = messages[i].From == username
	}
	return messages
}
