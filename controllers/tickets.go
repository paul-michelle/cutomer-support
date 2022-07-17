package controllers

import (
	"db-queries/db"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

type TicketDetails struct {
	Topic string `json:"topic"`
	Text  string `json:"text"`
}

// Methods: GET/POST; path: /tickets

func (h *BaseHandler) TicketsListAllOrCreateOne(w http.ResponseWriter, authReq *AuthenticatedRequest) {
	switch authReq.Method {
	case "GET":
		h.GetAllTickets(w, authReq)
	case "POST":
		h.CreateTicket(w, authReq)
	default:
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
	}
}

func (h *BaseHandler) GetAllTickets(w http.ResponseWriter, authReq *AuthenticatedRequest) {
	tickets, err := db.GetTicketsForUser(h.Conn, authReq.user.Email, authReq.user.IsStaff, authReq.user.IsSuperuser)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets)
}

func (h *BaseHandler) CreateTicket(w http.ResponseWriter, authReq *AuthenticatedRequest) {
	if authReq.Body == nil {
		http.Error(w, "Payload expected.", http.StatusBadRequest)
		return
	}

	var ticket TicketDetails
	err := json.NewDecoder(authReq.Body).Decode(&ticket)
	if err != nil {
		http.Error(w, "Invalid payload.", http.StatusBadRequest)
		return
	}

	if ticket.Topic == "" || ticket.Text == "" {
		http.Error(w, "Missing fields in payload: expected topic and text.", http.StatusBadRequest)
		return
	}

	id, err := db.CreateTicket(h.Conn, authReq.user.Email, ticket.Topic, ticket.Text)
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code.Name() == db.VALUE_TOO_LONG_ERR_CODE_NAME {
			http.Error(w, "Provided values are exceeding max chars limit.", http.StatusBadRequest)
			return
		}
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := make(map[string]int)
	resp["id"] = id
	json.NewEncoder(w).Encode(resp)
}

// Methods: GET/PUT/PATCH; path: /tickets/{id}

func (h *BaseHandler) TicketsGetOrUpdateOne(w http.ResponseWriter, authReq *AuthenticatedRequest) {
	idInQuery := strings.TrimPrefix(authReq.URL.Path, "/tickets/")
	switch authReq.Method {
	case "GET":
		h.GetOneTicket(idInQuery, w, authReq)
	default:
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
	}
}

func (h *BaseHandler) GetOneTicket(id string, w http.ResponseWriter, authReq *AuthenticatedRequest) {
	ticket, err := db.GetOneTicketForUser(h.Conn, id, authReq.user.Email, authReq.user.IsStaff, authReq.user.IsSuperuser)
	if err != nil {
		http.Error(w, "Ticket does not exist or does not belong to this user.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}
