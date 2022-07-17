package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"db-queries/controllers"
	"db-queries/db"

	"github.com/joho/godotenv"
)

func GetEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to parse .env file.")
	}
	var ()
	log.Println("Initializing DB connection.")
	conn, err := db.Initialize(
		&db.DSN{
			HOST:     os.Getenv("DB_HOST"),
			PORT:     os.Getenv("DB_PORT"),
			USERNAME: os.Getenv("DB_USER"),
			PASSWORD: os.Getenv("DB_PASSWORD"),
			DATABASE: os.Getenv("DB_NAME"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating relations in DB.")
	err = db.CreateRelations(conn)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Registering routes.")
	h := controllers.NewBaseHandler(conn)
	http.HandleFunc("/time", h.Pong)
	http.HandleFunc("/users", h.UsersListAllOrCreateOne)
	http.HandleFunc("/login", h.LogIn)
	http.Handle("/tickets", controllers.JWTMiddleWare(h.TicketsListAllOrCreateOne))
	http.Handle("/tickets/", controllers.JWTMiddleWare(h.TicketsDetailedView))

	log.Println("Initializing HTTP server.")
	host := GetEnv("SERVER_HOST", "127.0.0.1")
	port := GetEnv("SERVER_PORT", "8089")
	servAddr := fmt.Sprintf("%s:%s", host, port)
	s := &http.Server{Addr: servAddr}

	log.Printf("Starting server at %s", servAddr)
	log.Fatal(s.ListenAndServe())

	defer h.Conn.Close()
	defer log.Println("Closing DB connection.")
}
