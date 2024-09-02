package chat

import (
	"log"
	"net/http"

	"github.com/aabuezo/go-simple-chat/config"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type TextMessage struct {
	From    string
	To      string
	Message string
}

var ActiveUsers = []config.User{}
var Messages = []config.Message{}
var Sessions = map[string]string{}

func GetHome(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func PostLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {

		username := req.FormValue("username")
		password := req.FormValue("password")

		log.Printf("loggin in as: %s\n", username)

		if alreadyLoggedIn(username) {
			http.Redirect(w, req, "/room", http.StatusSeeOther)
		}

		if !Authenticate(username, password) {
			http.Error(w, "Invalid username or password", http.StatusBadRequest)
			return
		}

		log.Println("PostLogin() username: ", username)

		// user := GetUser(username)
		// now the user is Active
		// ActiveUsers = append(ActiveUsers, user)

		c, err := req.Cookie("chat_sid")
		if err != nil {
			sID := uuid.NewV4()
			c = &http.Cookie{
				Name:  "chat_sid",
				Value: sID.String(),
			}
			Sessions[sID.String()] = username
			log.Printf("PostLogin: username %s added to Sessions\n", Sessions[sID.String()])
		}
		http.SetCookie(w, c)

		// redirect to chat room
		http.Redirect(w, req, "/room", http.StatusFound)
		// config.TPL.ExecuteTemplate(w, "chat-room.htm", nil)
		return
	}
	config.TPL.ExecuteTemplate(w, "login.htm", nil)
}

func alreadyLoggedIn(username string) bool {
	for _, user := range ActiveUsers {
		log.Printf("user.Username: %s\n", user.Username)
		if user.Username == username {
			return true
		}
	}
	return false
}

// Authenticate searches for the user in the DB and verfies that the provided
// username and password are valid
func Authenticate(username, password string) bool {
	if username == "" || password == "" {
		return false
	}

	dbuser := GetUser(username)
	if dbuser.ID == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword(dbuser.Password, []byte(password))
	log.Println(err)
	return err == nil
}

func GetChats(w http.ResponseWriter, req *http.Request) {
	if IsOpenSession(req) {
		sID, err := req.Cookie("chat_sid")
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusForbidden)
			return
		}

		username := Sessions[sID.Value]
		log.Println("in GetChats() username: ", username)
		user := GetUser(username)
		messages := GetMessages(user, user)

		textMessages := messageIDsToText(messages)

		config.TPL.ExecuteTemplate(w, "chats.htm", textMessages)
	}
}

func GetChatRoom(w http.ResponseWriter, req *http.Request) {
	if IsOpenSession(req) {
		sID, err := req.Cookie("chat_sid")
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusForbidden)
			return
		}

		// username := Sessions[strings.Split(sID.String(), "=")[1]] // Aca estaba el problema!!!
		username := Sessions[sID.Value] // Aca estaba el problema!!!

		config.TPL.ExecuteTemplate(w, "chat-room.htm", username)
	}
}

func IsOpenSession(req *http.Request) bool {
	_, err := req.Cookie("chat_sid")
	return err == nil
}

func PostMessage(w http.ResponseWriter, req *http.Request) {

	if IsOpenSession(req) {
		message := req.FormValue("message")

		log.Println("in PostMessage() message: ", message)

		sID, err := req.Cookie("chat_sid")
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// from := Sessions[strings.Split(sID.String(), "=")[1]]
		from := Sessions[sID.Value]
		log.Printf("PostMessage() from: %s\n", from)

		to := req.FormValue("to")
		log.Println("in PostMessage() to: ", to)

		fromUser := GetUser(from)
		toUser := GetUser(to)
		SaveMessage(fromUser, toUser, message)
		// http.Redirect(w, req, "/room", http.StatusSeeOther)
	}

	http.Redirect(w, req, "/room", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, req *http.Request) {
	c, _ := req.Cookie("chat_sid")
	// delete the session
	delete(Sessions, c.Value)
	// remove the cookie
	c = &http.Cookie{
		Name:   "chat_sid",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)

	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func messageIDsToText(messages []config.Message) []TextMessage {
	users := GetUsers()
	usersMap := map[int]string{}

	for _, user := range users {
		usersMap[user.ID] = user.Username
	}

	textMessage := TextMessage{}
	textMessages := []TextMessage{}
	for _, message := range messages {
		textMessage.From = usersMap[message.From]
		textMessage.To = usersMap[message.To]
		textMessage.Message = message.Message
		textMessages = append(textMessages, textMessage)
	}
	return textMessages
}
