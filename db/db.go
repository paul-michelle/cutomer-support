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

const (
	enableCryptoStmt = "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
	createTableUsersStmt = `
CREATE TABLE IF NOT EXISTS users
(
    id SERIAL,
    created_at TIMESTAMP DEFAULT now(),
	email VARCHAR(64) NOT NULL UNIQUE,
    password TEXT NOT NULL,
	username VARCHAR(64) NOT NULL,
	is_staff BOOLEAN DEFAULT FALSE,
	is_superuser BOOLEAN DEFAULT FALSE,
    CONSTRAINT pk_users PRIMARY KEY (id)
);`
	createStatusTypeStmt = `
CREATE OR REPLACE FUNCTION create_ticket_status_type() RETURNS integer AS $$
DECLARE type_already_exists INTEGER;
	BEGIN
		SELECT into type_already_exists (SELECT 1 FROM pg_type WHERE typname = 'status');
		IF type_already_exists IS NULL THEN
			CREATE TYPE status AS ENUM ('resolved', 'unresolved', 'pending', 'unknown');
		END IF;
		RETURN type_already_exists;
	END;
	$$ LANGUAGE plpgsql;
SELECT create_ticket_status_type();
DROP function create_ticket_status_type();`

	createTableTicketsStmt = `
CREATE TABLE IF NOT EXISTS tickets
(
    id SERIAL,
    created_at TIMESTAMP DEFAULT now(),
	updated_at TIMESTAMP DEFAULT now(),
	author INTEGER REFERENCES users (id),
    topic VARCHAR(20) NOT NULL,
	status STATUS,
    CONSTRAINT pk_tickets PRIMARY KEY (id)
);`
	createTableMessagesStmt = `
CREATE TABLE IF NOT EXISTS messages
(
    id SERIAL,
    created_at TIMESTAMP DEFAULT now(),
	author INTEGER REFERENCES users (id),
	text TEXT,
	ticket INTEGER REFERENCES tickets (id),
	CONSTRAINT pk_messages PRIMARY KEY (id)
);`
	createTicketStmt	= "INSERT INTO tickets (customer, topic, contents) VALUES ($1, $2, $3) RETURNING id"
	getAllTicketsStmt	= "SELECT * FROM tickets ORDER BY created_at ASC"
	createUserStmt		= `
	INSERT INTO users (email, password, username, is_staff, is_superuser) 
	VALUES ($1, crypt($2, gen_salt('bf', 8)), $3, $4, $5) RETURNING id;`
	checkUserExistsStmt = "SELECT exists(SELECT 1 FROM users WHERE email=$1 and password=crypt($2, password));"
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

func CreateRelations(conn *sql.DB) (err error) {
	_, err = conn.Exec(enableCryptoStmt)
	if err != nil {
		return err
	}

	log.Println("Creating table 'users' if not exists.")
	_, err = conn.Exec(createTableUsersStmt)
	if err != nil {
		return err
	}

	_, err = conn.Exec(createStatusTypeStmt)
	if err != nil {
		return err
	}

	log.Println("Creating table 'tickets' if not exists.")
	_, err = conn.Exec(createTableTicketsStmt)
	if err != nil {
		return err
	}

	log.Println("Creating table 'messages' if not exists.")
	_, err = conn.Exec(createTableMessagesStmt)
	if err != nil {
		return err
	}
	return nil
}

func CreateTicket(conn *sql.DB, customer, topic, contents string) (lastInsertId int, err error) {
	err = conn.QueryRow(createTicketStmt, customer, topic, contents).Scan(&lastInsertId)
	return lastInsertId, err
}

func GetAllTickets(conn *sql.DB) (tickets TicketsList, err error) {
	rows, err := conn.Query(getAllTicketsStmt)
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

func CreateUser(conn *sql.DB, email, password, username string,
	is_staff, is_superuser bool) (lastInsertId int, err error) {
	err = conn.QueryRow(createUserStmt, email, password, username, is_staff, is_superuser).Scan(&lastInsertId)
	return lastInsertId, err
}

func UserExists(conn *sql.DB, email, password string) (exists bool, err error) {
	err = conn.QueryRow(checkUserExistsStmt, email, password).Scan(&exists)
	return exists, err
}