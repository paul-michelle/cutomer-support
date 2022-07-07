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

type Database struct {
	Conn *sql.DB
}

func Initialize(username, password, database string) (Database, error) {
	db := Database{}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		HOST, PORT, USERNAME, PASSWORD, DATABASE)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return db, err
	}

	db.Conn = conn
	err = db.Conn.Ping()
	if err != nil {
		return db, err
	}

	log.Println("Database connection established")
	return db, nil
}

func main() {
	db, err := Initialize(os.Getenv("username"),
		os.Getenv("password"),
		os.Getenv("database"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(db)
}