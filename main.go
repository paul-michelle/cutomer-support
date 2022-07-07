package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var (
	HOST     = os.Getenv("DB_HOST")
	PORT     = os.Getenv("DB_PORT")
	USERNAME = os.Getenv("DB_USER")
	PASSWORD = os.Getenv("DB_PASSWORD")
	DATABASE = os.Getenv("DB_NAME")
)

var (
	SERVER_HOST = os.Getenv("SERVER_HOST")
	SERVER_PORT = os.Getenv("SERVER_PORT")
)

type BaseHandler struct {
	Conn *sql.DB
}

func (h *BaseHandler) Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{ db }
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS tickets
(
    id SERIAL,
    created_at TIMESTAMP DEFAULT now(),
	updated_at TIMESTAMP DEFAULT now(),
	customer VARCHAR(20) NOT NULL,
    topic VARCHAR(20) NOT NULL,
	contents TEXT NOT NULL,
    CONSTRAINT pk_tickets PRIMARY KEY (id)
)`

func Initialize() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		HOST, PORT, USERNAME, PASSWORD, DATABASE)
	conn, err := sql.Open("postgres", dsn)
	if err != nil { return nil, err }
	if err := conn.Ping(); err != nil { return nil, err }
	log.Println("Database connection established.")

	return conn, nil
}

func main() {
	log.Println("Initializing DB connection.")
	conn, err := Initialize()
	if err != nil { log.Fatal(err) }
	
	h := NewBaseHandler(conn)

	log.Println("Creating tickets relation in DB.")
	_, err = h.Conn.Exec(tableCreationQuery)
	if err != nil { log.Fatal(err) }
	
	log.Println("Registering routes.")
	http.HandleFunc("/time", h.Pong)

	s := &http.Server{
		Addr: fmt.Sprintf("%s:%s", SERVER_HOST, SERVER_PORT),
	}
	log.Printf("Starting server on port %s", SERVER_PORT)
	s.ListenAndServe()

	defer h.Conn.Close()
	defer log.Println("Closing DB connection.")
}