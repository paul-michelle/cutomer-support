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
		log.Println("WARNING. Unable to parse .env file.")
	}

	log.Println("Initializing DB connection.")
	conn, err := db.Initialize(
		&db.DSN{
			HOST:     GetEnv("DB_HOST", "127.0.0.1"),
			PORT:     GetEnv("DB_PORT", "5433"),
			USERNAME: GetEnv("DB_USER", "postgres"),
			PASSWORD: GetEnv("DB_PASSWORD", "postgres"),
			DATABASE: GetEnv("DB_NAME", "tickets"),
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
