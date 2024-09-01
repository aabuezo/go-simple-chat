package main

import (
	"net/http"

	"github.com/aabuezo/go-simple-chat/chat"
)

func main() {

	http.HandleFunc("/", chat.PostLogin)
	http.HandleFunc("/chat", chat.PostMessage)

	http.ListenAndServe(":8090", nil)
}
