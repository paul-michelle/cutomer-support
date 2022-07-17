package controllers

import (
	"db-queries/db"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

const ID_POSITION_IN_URL_PATH = 2

var (
	ticketOperationRegex, _ = regexp.Compile("^/tickets/[0-9]+[/]?$")
	msgOperationRegex, _    = regexp.Compile("^/tickets/[0-9]+/messages[/]?$")
)

type TicketDetails struct {
	Topic  string `json:"topic"`
	Text   string `json:"text"`
	Status string `json:"status,omitempty"`
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

func (h *BaseHandler) TicketsDetailedView(res http.ResponseWriter, authReq *AuthenticatedRequest) {
	// Methods: GET/PUT/PATCH; path: /tickets/{id}
	if ticketOperationRegex.MatchString(authReq.URL.Path) {
		ticketId := strings.Split(authReq.URL.Path, "/")[ID_POSITION_IN_URL_PATH]
		switch {
		case authReq.Method == "GET":
			h.GetOneTicket(ticketId, res, authReq)

		case authReq.Method == "PUT" || authReq.Method == "PATCH":
			var ticket TicketDetails
			err := json.NewDecoder(authReq.Body).Decode(&ticket)
			if err != nil {
				http.Error(res, "Invalid payload.", http.StatusBadRequest)
				return
			}
			validChangeByStaff := (authReq.user.IsStaff || authReq.user.IsSuperuser) && db.VALID_TICKET_STATUS_STAFF[ticket.Status]
			validChangeByCommonUser := !(authReq.user.IsStaff || authReq.user.IsSuperuser) && db.VALID_TICKET_STATUS_COMMON_USER[ticket.Status]
			if validChangeByStaff || validChangeByCommonUser {
				h.UpdateTicket(ticketId, ticket.Status, res, authReq)
				return
			}
			http.Error(res, "Invalid status.", http.StatusBadRequest)

		default:
			http.Error(res, "Method Not Allowed.", http.StatusMethodNotAllowed)
		}
		return
	}
	// Methods: GET/POST; path /tickets/{id}/messages
	if msgOperationRegex.MatchString(authReq.URL.Path) {
		ticketId := strings.Split(authReq.URL.Path, "/")[ID_POSITION_IN_URL_PATH]
		switch {
		case authReq.Method == "GET":
			fmt.Println(ticketId)
			return
		case authReq.Method == "POST":
			return
		default:
			http.Error(res, "Method Not Allowed.", http.StatusMethodNotAllowed)
		}
		return
	}
	http.Error(res, "", http.StatusBadRequest)
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

func (h *BaseHandler) UpdateTicket(id, status string, res http.ResponseWriter, authReq *AuthenticatedRequest) {
	if !db.UpdateTicket(h.Conn, id, status) {
		http.Error(res, "Ticket does not exist or does not belong to this user.", http.StatusNotFound)
		return
	}
}
