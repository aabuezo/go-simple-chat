package chat

import (
	"log"
	"time"

	"github.com/aabuezo/go-simple-chat/config"
)

type User struct {
	ID       int
	Name     string
	Email    string
	Password []byte
}

type Message struct {
	ID      int
	Posted  time.Time
	From    int
	To      int
	Message string
}

func CreateTableUsers() {

	query := `CREATE TABLE IF NOT EXISTS users (
		user_id  INT GENERATED ALWAYS AS IDENTITY,
		name     VARCHAR(50) NOT NULL,
		email    VARCHAR(50) NOT NULL,
		password VARCHAR(50) NOT NULL,
		PRIMARY KEY(user_id)
	);`
	_, err := config.DB.Query(query)
	if err != nil {
		log.Fatalln("Could not create `users` table.")
		panic(err)
	}
}

func CreateTableMessages() {

	query := `CREATE TABLE IF NOT EXISTS messages (
		message_id  INT GENERATED ALWAYS AS IDENTITY,
		posted      TIMESTAMPTZ,
		from_id     INT    NOT NULL,
		to_id       INT    NOT NULL,
		message     VARCHAR(255) NOT NULL,
		PRIMARY KEY(message_id),
		CONSTRAINT fk_from
			FOREIGN KEY(from_id) 
			REFERENCES users(user_id),
		CONSTRAINT fk_to
			FOREIGN KEY(to_id) 
			REFERENCES users(user_id)
		);`
	_, err := config.DB.Query(query)
	if err != nil {
		log.Fatalln("Could not create `messages` table.")
		panic(err)
	}
}

func CreateUsers() {

	query := `INSERT INTO users (name, email, password) 
	VALUES 
		('John', 'john@mail.com', 'test'),
		('Barney', 'barney@mail.com', 'test'),
		('Anna', 'anna@mail.com', 'test'),
		('Janeth', 'janeth@mail.com', 'test'),
		('Luka', 'luka@mail.com', 'test'),
		('Stacey', 'stacey@mail.com', 'test')`
	_, err := config.DB.Query(query)
	if err != nil {
		log.Fatalln("Could not populate table `users`.")
		panic(err)
	}
}

func GetUsers() {

}
