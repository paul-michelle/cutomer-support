package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"db-queries/controllers"
	"db-queries/db"
)

var (
	SERVER_HOST = os.Getenv("SERVER_HOST")
	SERVER_PORT = os.Getenv("SERVER_PORT")
)

func main() {
	log.Println("Initializing DB connection.")
	conn, err := db.Initialize()
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
	http.HandleFunc("/tickets", h.TicketsListAllOrCreateOne)

	log.Printf("Starting server at %s on port %s", SERVER_HOST, SERVER_PORT)
	s := &http.Server{Addr: fmt.Sprintf("%s:%s", SERVER_HOST, SERVER_PORT)}
	log.Fatal(s.ListenAndServe())

	defer h.Conn.Close()
	defer log.Println("Closing DB connection.")
}
