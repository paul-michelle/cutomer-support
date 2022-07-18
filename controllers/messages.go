package controllers

import (
	"db-queries/db"
	"encoding/json"
	"net/http"
)

func (h *BaseHandler) GetMessagesForTicket(ticketId string, res http.ResponseWriter, authReq *AuthenticatedRequest) {
	msgs, err := db.GetMessagesForTicket(h.Conn, ticketId)
	if err != nil || msgs == nil {
		http.Error(res, "No messages found.", http.StatusNotFound)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(msgs)
}

func (h *BaseHandler) CreateMessage(ticketID string, res http.ResponseWriter, authReq *AuthenticatedRequest) {
	var details TicketDetails
	err := json.NewDecoder(authReq.Body).Decode(&details)
	if err != nil || details.Text == "" {
		http.Error(res, "Text of the message expected.", http.StatusBadRequest)
		return
	}

	msgType := "request"
	if authReq.user.IsStaff || authReq.user.IsSuperuser {
		msgType = "response"
	}

	if !db.AddMessage(h.Conn, msgType, authReq.user.Email, ticketID, details.Text) {
		http.Error(res, "Ticket does not exist or does not belong to this user.", http.StatusNotFound)
		return
	}
	res.WriteHeader(http.StatusCreated)
}
