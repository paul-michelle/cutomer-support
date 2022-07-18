package db

import (
	"database/sql"
	"time"
)

const (
	GET_MESSAGES_FOR_TICKET_STMT = "SELECT created_at, type, text FROM messages WHERE ticket=$1"
	ADD_MESSAGE_TO_TICKET_STMT   = "INSERT INTO messages (type, author, text, ticket) VALUES ($1, $2, $3, $4)"
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

func AddMessage(conn *sql.DB, msgType, author, ticketId, text string) bool {
	exeResults, err := conn.Exec(ADD_MESSAGE_TO_TICKET_STMT, msgType, author, text, ticketId)
	if err != nil {
		return false
	}

	if rowsAffected, _ := exeResults.RowsAffected(); rowsAffected == 0 {
		return false
	}

	return true
}
