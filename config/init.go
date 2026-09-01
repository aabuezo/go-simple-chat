package config

import (
	"database/sql"
	"html/template"
	"io/fs"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB
var TPL *template.Template
var InitialData = false

type User struct {
	ID       int
	Username string
	Password []byte
}

type Message struct {
	ID      int
	From    int
	To      int
	Message string
}

// CREATE USER golang WITH PASSWORD 'password';
// ALTER USER golang WITH SUPERUSER;
// GRANT ALL PRIVILEGES ON DATABASE chat to golang;

func InitDB() {
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		db, openErr := sql.Open("postgres", "postgres://postgres:postgres@db:5432/chat?sslmode=disable")
		// db, openErr := sql.Open("postgres", "postgres://postgres:postgres@localhost/chat?sslmode=disable")
		if openErr == nil {
			if err = db.Ping(); err == nil {
				DB = db
				break
			}
			db.Close()
		} else {
			err = openErr
		}
		log.Printf("database is not ready (attempt %d/30): %v", attempt, err)
		time.Sleep(time.Second)
	}
	if DB == nil {
		panic(err)
	}

	CreateTableUsers()
	CreateTableMessages()
	CreateUsers()
	// CreateMessages()
}

func InitTemplates(files fs.FS) {
	TPL = template.Must(template.ParseFS(files, "templates/*"))
}

func CreateTableUsers() {

	query := `CREATE TABLE IF NOT EXISTS users (
		user_id  INT GENERATED ALWAYS AS IDENTITY,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		PRIMARY KEY(user_id)
	);`
	_, err := DB.Query(query)
	if err != nil {
		log.Fatalln("Could not create `users` table.")
		panic(err)
	}
	log.Println("Created table `users`")
}

func CreateTableMessages() {

	query := `CREATE TABLE IF NOT EXISTS messages (
		message_id  INT GENERATED ALWAYS AS IDENTITY,
		from_id     INT    NOT NULL,
		to_id       INT    NOT NULL,
		message     TEXT NOT NULL,
		PRIMARY KEY(message_id),
		CONSTRAINT fk_from
			FOREIGN KEY(from_id) 
			REFERENCES users(user_id),
		CONSTRAINT fk_to
			FOREIGN KEY(to_id) 
			REFERENCES users(user_id)
		);`
	_, err := DB.Query(query)
	if err != nil {
		log.Fatalln("Could not create `messages` table.")
		panic(err)
	}
	log.Println("Created table `messages`")
}

func CreateUsers() {

	bs, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	log.Println(string(bs))
	query := `INSERT INTO users (username, password)
	SELECT username, $1
	FROM (VALUES ('John'), ('Barney'), ('Anna'), ('Janeth'), ('Luka'), ('Stacey')) AS default_users(username)
	WHERE NOT EXISTS (
		SELECT 1 FROM users WHERE users.username = default_users.username
	)`
	_, err := DB.Exec(query, string(bs))
	if err != nil {
		log.Fatalln("Could not populate table `users`.")
		panic(err)
	}
	log.Println("Created users for testing")
}

func CreateMessages() {

	query := `INSERT INTO messages (from_id, to_id, message) VALUES (2, 1, 'Hello John!'), (3, 1, 'Hi, there'), (1, 2, 'Hi, Im John'), (3, 2, 'Hello Barney');`
	_, err := DB.Query(query)
	if err != nil {
		log.Fatalln("Could not populate table `messages`.")
		panic(err)
	}
	log.Println("Created dummy messages")
}
