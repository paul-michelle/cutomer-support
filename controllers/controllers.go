package controllers

import (
	"database/sql"
	"db-queries/db"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type BaseHandler struct {
	Conn *sql.DB
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{db}
}

type TicketDetails struct {
	Customer string `json: "customer"`
	Topic    string `json: "topic"`
	Contents string `json: "contents"`
}

func (h *BaseHandler) Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}

func (h *BaseHandler) TicketsListAllOrCreateOne(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.GetAllTickets(w, r)
		return
	}
	if r.Method == "POST" {
		h.CreateTicket(w, r)
		return
	}
	http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
}

func (h *BaseHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var ticket TicketDetails
	if r.Body == nil {
		http.Error(w, "Payload expected.", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&ticket)
	if err != nil {
		http.Error(w, "Invalid payload.", http.StatusBadRequest)
		return
	}

	if ticket.Customer == "" || ticket.Topic == "" || ticket.Contents == "" {
		http.Error(w, "Missing fields in payload", http.StatusBadRequest)
		return
	}

	id, err := db.CreateTicket(h.Conn, ticket.Customer, ticket.Topic, ticket.Contents)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := make(map[string]int)
	resp["id"] = id
	json.NewEncoder(w).Encode(resp)
}

func (h *BaseHandler) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := db.GetAllTickets(h.Conn)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets.Tickets)
}
