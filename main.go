package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	HOST     = os.Getenv("DB_HOST")
	PORT     = os.Getenv("DB_PORT")
	USERNAME = os.Getenv("DB_USER")
	PASSWORD = os.Getenv("DB_PASSWORD")
	DATABASE = os.Getenv("DB_NAME")
)

type BaseHandler struct {
	Conn *sql.DB
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
	log.Println("Database connection established")
	return conn, nil
}

func main() {
	conn, err := Initialize()
	if err != nil { log.Fatal(err) }
	
	h := NewBaseHandler(conn)

	_, err = h.Conn.Exec(tableCreationQuery)
	if err != nil { log.Fatal(err) }
	
	defer h.Conn.Close()
	defer fmt.Println("Closing DB connection.")
}