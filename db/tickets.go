package db

import (
	"database/sql"
	"time"
)

const (
	DEFAULT_TICKET_STATUS         = "pending"
	GET_TICKETS_OF_THIS_USER_STMT = "SELECT * FROM tickets WHERE author=$1 ORDER BY created_at ASC"
	GET_ALL_TICKETS_STMT          = "SELECT * FROM tickets ORDER BY created_at ASC"
	CREATE_TICKET_STMT            = `
	WITH insert_to_tickets AS 
	(INSERT INTO tickets (author, topic, status) 
	VALUES ($1, $2, $3) RETURNING id)
    INSERT INTO messages (author, ticket, text) 
	VALUES ($1, (SELECT id FROM insert_to_tickets), $4) RETURNING id;
`
)

type Ticket struct {
	ID     int
	CrtdAt time.Time
	UpdAt  time.Time
	Author string
	Topic  string
	Status string
}

func CreateTicket(conn *sql.DB, email, topic, text string) (lastInsertId int, err error) {
	err = conn.QueryRow(CREATE_TICKET_STMT, email, topic, DEFAULT_TICKET_STATUS, text).Scan(&lastInsertId)
	return lastInsertId, err
}

func GetTicketsForUser(conn *sql.DB, email string, isStaff, isSuperuser bool) (tickets []Ticket, err error) {
	var rows *sql.Rows
	switch {
	case isStaff || isSuperuser:
		rows, err = conn.Query(GET_ALL_TICKETS_STMT)
		if err != nil {
			return tickets, err
		}
	default:
		rows, err = conn.Query(GET_TICKETS_OF_THIS_USER_STMT, email)
		if err != nil {
			return tickets, err
		}
	}

	for rows.Next() {
		var ticket Ticket
		err := rows.Scan(&ticket.ID, &ticket.CrtdAt, &ticket.UpdAt, &ticket.Author, &ticket.Topic, &ticket.Status)
		if err != nil {
			return tickets, err
		}

		tickets = append(tickets, ticket)
	}
	return tickets, nil
}
