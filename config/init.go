package config

import (
	"database/sql"
	"html/template"

	_ "github.com/lib/pq"

	"github.com/aabuezo/go-simple-chat/chat"
)

var DB *sql.DB
var TPL *template.Template
var Users []chat.User
var Messages []chat.Message

func init() {
	var err error
	DB, err = sql.Open("postgres", "postgres://postgres.postgres@localhost/chat?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = DB.Ping(); err != nil {
		panic(err)
	}

	chat.CreateTableUsers()
	chat.CreateTableMessages()

	Users = []chat.User{}
	Messages = []chat.Message{}

	TPL = template.Must(template.ParseGlob("templates/*"))
}
