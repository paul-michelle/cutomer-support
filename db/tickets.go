package db

import (
	"database/sql"
	"time"
)

const (
	DEFAULT_TICKET_STATUS         = "pending"
	DEFAULT_MSG_TYPE              = "request"
	GET_ALL_TICKETS_STMT          = "SELECT id, created_at, updated_at, author, topic, status FROM tickets ORDER BY created_at ASC"
	GET_TICKETS_OF_THIS_USER_STMT = "SELECT id, created_at, updated_at, topic, status FROM tickets WHERE author=$1 ORDER BY created_at ASC"
	CREATE_TICKET_STMT            = `
	WITH insert_to_tickets AS 
	(INSERT INTO tickets (author, topic, status) 
	VALUES ($1, $2, $3) RETURNING id)
    INSERT INTO messages (author, ticket, text, type) 
	VALUES ($1, (SELECT id FROM insert_to_tickets), $4, $5) RETURNING id;
`
	GET_TICKET_STMT              = "SELECT created_at, updated_at, author, topic, status FROM tickets WHERE id=$1"
	GET_TICKET_OF_THIS_USER_STMT = "SELECT created_at, updated_at, topic, status FROM tickets WHERE id=$1 and author=$2"
	UPDATE_TICKET_STMT           = "UPDATE tickets SET status=$2 WHERE id=$1"
)

type Ticket struct {
	ID     int       `json:"id,omitempty"`
	CrtdAt time.Time `json:"created_at"`
	UpdAt  time.Time `json:"updated_at"`
	Author string    `json:"author,omitempty"`
	Topic  string    `json:"topic"`
	Status string    `json:"status"`
}

func CreateTicket(conn *sql.DB, email, topic, text string) (lastInsertId int, err error) {
	err = conn.QueryRow(CREATE_TICKET_STMT, email, topic, DEFAULT_TICKET_STATUS, text, DEFAULT_MSG_TYPE).Scan(&lastInsertId)
	return lastInsertId, err
}

func GetTicketsForUser(conn *sql.DB, email string, isStaff, isSuperuser bool) (tickets []Ticket, err error) {
	var rows *sql.Rows
	switch {
	case isStaff || isSuperuser:
		rows, err = conn.Query(GET_ALL_TICKETS_STMT)
	default:
		rows, err = conn.Query(GET_TICKETS_OF_THIS_USER_STMT, email)
	}

	if err != nil {
		return tickets, err
	}

	for rows.Next() {
		var ticket Ticket
		switch {
		case isStaff || isSuperuser:
			err = rows.Scan(&ticket.ID, &ticket.CrtdAt, &ticket.UpdAt, &ticket.Author, &ticket.Topic, &ticket.Status)
		default:
			err = rows.Scan(&ticket.ID, &ticket.CrtdAt, &ticket.UpdAt, &ticket.Topic, &ticket.Status)
		}

		if err != nil {
			return tickets, err
		}

		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func GetOneTicketForUser(conn *sql.DB, id, email string, isStaff, isSuperuser bool) (ticket Ticket, err error) {
	switch {
	case isStaff || isSuperuser:
		err = conn.QueryRow(GET_TICKET_STMT, id).Scan(&ticket.CrtdAt, &ticket.UpdAt, &ticket.Author, &ticket.Topic, &ticket.Status)
	default:
		err = conn.QueryRow(GET_TICKET_OF_THIS_USER_STMT, id, email).Scan(&ticket.CrtdAt, &ticket.UpdAt, &ticket.Topic, &ticket.Status)
	}
	return ticket, err
}

func UpdateTicket(conn *sql.DB, id, status string) bool {
	exeResults, err := conn.Exec(UPDATE_TICKET_STMT, id, status)
	if err != nil {
		return false
	}

	if rowsAffected, _ := exeResults.RowsAffected(); rowsAffected == 0 {
		return false
	}

	return true
}
