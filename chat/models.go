package chat

import (
	"log"

	"github.com/aabuezo/go-simple-chat/config"
)

func GetUser(username string) config.User {
	log.Printf("in GetUser(%s)\n", username)
	rows, err := config.DB.Query(`SELECT * FROM users WHERE username LIKE $1;`, username)
	if err != nil {
		panic(err)
	}
	dbUser := config.User{}
	for rows.Next() {
		rows.Scan(&dbUser.ID, &dbUser.Username, &dbUser.Password)
		// log.Printf("dbUser.ID: %d, dbUser.Username: %s, dbUser.Password: %s\n", dbUser.ID, dbUser.Username, dbUser.Password)
	}
	return dbUser // if not in DB returns zero value
}

func SaveMessage(from config.User, to config.User, message string) error {
	query := `INSERT INTO messages (from_id, to_id, message)
	VALUES ($1, $2, $3)`
	_, err := config.DB.Query(query, from.ID, to.ID, message)
	if err != nil {
		log.Println("Error when saving message to the DB")
		return err
	}
	log.Println("Message saved to the DB")
	return nil
}

func GetMessages(from config.User, to config.User) []config.Message {

	log.Printf("GetMessages(%s)\n", to.Username)

	query := `SELECT * FROM messages WHERE from_id=$1 OR to_id=$2;`
	rows, err := config.DB.Query(query, from.ID, to.ID)
	if err != nil {
		log.Println("Error when searching for messages")
		panic(err)
	}
	var messages = []config.Message{}
	for rows.Next() {
		message := config.Message{}
		err := rows.Scan(&message.ID, &message.From, &message.To, &message.Message)
		if err != nil {
			break
		}

		log.Printf("From: %d To: %d -> Message: %s\n", message.From, message.To, message.Message)

		messages = append(messages, message)
	}
	return messages
}

func GetUsers() []config.User {
	rows, err := config.DB.Query(`SELECT * FROM users;`)
	if err != nil {
		panic(err)
	}
	users := []config.User{}
	for rows.Next() {
		user := config.User{}
		rows.Scan(&user.ID, &user.Username, &user.Password)
		users = append(users, user)
	}
	return users
}
