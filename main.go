package main

import (
	"embed"
	"net/http"

	"github.com/aabuezo/go-simple-chat/chat"
	"github.com/aabuezo/go-simple-chat/config"
)

// templateFiles contains the HTML templates inside the application binary.
//
//go:embed templates/*
var templateFiles embed.FS

func main() {
	config.InitTemplates(templateFiles)

	http.HandleFunc("/", chat.GetHome)
	http.HandleFunc("/login", chat.PostLogin)
	http.HandleFunc("/logout", chat.Logout)
	http.HandleFunc("/room", chat.GetChatRoom)
	http.HandleFunc("/room/ws", chat.HandleWebSocket)
	http.HandleFunc("/room/message", chat.PostMessage)
	http.HandleFunc("/room/messages", chat.GetChats)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.ListenAndServe(":8090", nil)
}
