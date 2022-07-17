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
