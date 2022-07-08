package controllers

import (
	"time"
	"database/sql"

	_ "github.com/lib/pq"
)

type BaseHandler struct {
	Conn *sql.DB
}

func (h *BaseHandler) Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{ db }
}