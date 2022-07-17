package db

import (
	"database/sql"
	"time"
)

const (
	GET_MESSAGES_FOR_TICKET_STMT = "SELECT created_at, type, text FROM messages WHERE ticket=$1"
)

type Message struct {
	ID     int       `json:"id,omitempty"`
	CrtdAt time.Time `json:"created_at"`
	Type   string    `json:"type"`
	Author string    `json:"author,omitempty"`
	Text   string    `json:"text"`
	Ticket int       `json:"ticket,omitempty"`
}

func GetMessagesForTicket(conn *sql.DB, ticketId string) ([]Message, error) {
	rows, err := conn.Query(GET_MESSAGES_FOR_TICKET_STMT, ticketId)
	if err != nil {
		return nil, err
	}

	var msgs []Message
	for rows.Next() {
		var msg Message
		err = rows.Scan(&msg.CrtdAt, &msg.Type, &msg.Text)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
