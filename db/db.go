package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	HOST     = os.Getenv("DB_HOST")
	PORT     = os.Getenv("DB_PORT")
	USERNAME = os.Getenv("DB_USER")
	PASSWORD = os.Getenv("DB_PASSWORD")
	DATABASE = os.Getenv("DB_NAME")
)

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS tickets
(
    id SERIAL,
    created_at TIMESTAMP DEFAULT now(),
	updated_at TIMESTAMP DEFAULT now(),
	customer VARCHAR(20) NOT NULL,
    topic VARCHAR(20) NOT NULL,
	contents TEXT NOT NULL,
    CONSTRAINT pk_tickets PRIMARY KEY (id)
)`

const (
	createTicketQuery  = "INSERT INTO tickets (customer, topic, contents) VALUES ($1, $2, $3) RETURNING id"
	getAllTicketsQuery = "SELECT * FROM tickets ORDER BY created_at ASC"
)

type Ticket struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Customer  string    `json:"customer"`
	Topic     string    `json:"topic"`
	Contents  string    `json:"contents"`
}
type TicketsList struct {
	Tickets []Ticket `json:"tickets"`
}

func Initialize() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		HOST, PORT, USERNAME, PASSWORD, DATABASE)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	log.Println("Database connection established.")

	return conn, nil
}

func CreateTables(conn *sql.DB) (err error) {
	_, err = conn.Exec(tableCreationQuery)
	return err
}

func CreateTicket(conn *sql.DB, customer, topic, contents string) (lastInsertId int, err error) {
	err = conn.QueryRow(createTicketQuery, customer, topic, contents).Scan(&lastInsertId)
	return lastInsertId, err
}

func GetAllTickets(conn *sql.DB) (tickets TicketsList, err error) {
	rows, err := conn.Query(getAllTicketsQuery)
	if err != nil {
		return tickets, err
	}

	for rows.Next() {
		var ticket Ticket
		err := rows.Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt,
			&ticket.Customer, &ticket.Topic, &ticket.Contents)
		if err != nil {
			return tickets, err
		}

		tickets.Tickets = append(tickets.Tickets, ticket)
	}
	return tickets, nil
}
