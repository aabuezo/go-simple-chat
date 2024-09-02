package main

import (
	"net/http"

	"github.com/aabuezo/go-simple-chat/chat"
)

func main() {

	http.HandleFunc("/", chat.GetHome)
	http.HandleFunc("/login", chat.PostLogin)
	http.HandleFunc("/logout", chat.Logout)
	http.HandleFunc("/room", chat.GetChatRoom)
	http.HandleFunc("/room/message", chat.PostMessage)
	http.HandleFunc("/room/messages", chat.GetChats)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.ListenAndServe(":80", nil)
}
