package controllers

import (
	"db-queries/db"
	"encoding/json"
	"net/http"
)

type TicketDetails struct {
	Author   int    `json:"author"`
	Topic    string `json:"topic"`
	Contents string `json:"contents"`
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

func (h *BaseHandler) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := db.GetAllTickets(h.Conn)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets.Tickets)
}

func (h *BaseHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Payload expected.", http.StatusBadRequest)
		return
	}

	var ticket TicketDetails
	err := json.NewDecoder(r.Body).Decode(&ticket)
	if err != nil {
		http.Error(w, "Invalid payload.", http.StatusBadRequest)
		return
	}

	if ticket.Author == 0 || ticket.Topic == "" || ticket.Contents == "" {
		http.Error(w, "Missing fields in payload", http.StatusBadRequest)
		return
	}

	id, err := db.CreateTicket(h.Conn, ticket.Author, ticket.Topic)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := make(map[string]int)
	resp["id"] = id
	json.NewEncoder(w).Encode(resp)
}
