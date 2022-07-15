package controllers

import (
	"database/sql"
	"net/http"
	"time"
)

type Requester struct {
	Username    string
	Email       string
	IsStaff     bool
	IsSuperuser bool
}
type AuthenticatedRequest struct {
	*http.Request
	user Requester
}

type BaseHandler struct {
	Conn *sql.DB
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{db}
}

func (h *BaseHandler) Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}
