package controllers

import (
	"database/sql"
	"net/http"
	"time"
)

type Requester struct {
	IsStaff, IsSuperuser bool
	Email, Username      string
}
type RichRequest struct {
	*http.Request
	requester Requester
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
