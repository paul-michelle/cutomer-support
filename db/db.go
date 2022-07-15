package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DSN struct {
	HOST, PORT, USERNAME, PASSWORD, DATABASE string
}

const (
	enableCryptoStmt     = "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
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
	author VARCHAR(64) REFERENCES users (email),
    topic VARCHAR(20) NOT NULL,
	status STATUS,
    CONSTRAINT pk_tickets PRIMARY KEY (id)
);`
	createTableMessagesStmt = `
CREATE TABLE IF NOT EXISTS messages
(
    id SERIAL,
    created_at TIMESTAMP DEFAULT now(),
	author VARCHAR(64) REFERENCES users (email),
	text TEXT,
	ticket INTEGER REFERENCES tickets (id),
	CONSTRAINT pk_messages PRIMARY KEY (id)
);`
	createTicketStmt          = "INSERT INTO tickets (author, topic, status) VALUES ($1, $2, $3) RETURNING id"
	getTicketsUserCreatedStmt = "SELECT * FROM tickets WHERE author=$1 ORDER BY created_at ASC"
	getAllTicketsStmt         = "SELECT * FROM tickets ORDER BY created_at ASC"
	createUserStmt            = `
	INSERT INTO users (email, password, username, is_staff, is_superuser) 
	VALUES ($1, crypt($2, gen_salt('bf', 8)), $3, $4, $5);`
	getUserDetailsStmt = `
	SELECT id, username, email, is_staff, is_superuser FROM  users WHERE email=$1 and password=crypt($2, password);`
)

type Ticket struct {
	ID     int
	CrtdAt time.Time
	UpdAt  time.Time
	Author string
	Topic  string
	Status string
}

type User struct {
	ID          int
	Username    string
	Email       string
	IsStaff     bool
	IsSuperuser bool
}

func Initialize(dsn *DSN) (*sql.DB, error) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dsn.HOST, dsn.PORT, dsn.USERNAME, dsn.PASSWORD, dsn.DATABASE)

	conn, err := sql.Open("postgres", connString)
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

func CreateTicket(conn *sql.DB, author int, topic string) (lastInsertId int, err error) {
	err = conn.QueryRow(createTicketStmt, author, topic, "pending").Scan(&lastInsertId)
	return lastInsertId, err
}

func GetTicketsForUser(conn *sql.DB, email string, isStaff, isSuperuser bool) (tickets []Ticket, err error) {
	var rows *sql.Rows
	switch {
	case isStaff || isSuperuser:
		rows, err = conn.Query(getAllTicketsStmt)
		if err != nil {
			return tickets, err
		}
	default:
		rows, err = conn.Query(getTicketsUserCreatedStmt, email)
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

func CreateUser(conn *sql.DB, email, password, username string,
	is_staff, is_superuser bool) error {
	_, err := conn.Exec(createUserStmt, email, password, username, is_staff, is_superuser)
	return err
}

func GetUserDetails(conn *sql.DB, email, password string) (user User, err error) {
	err = conn.QueryRow(getUserDetailsStmt, email, password).Scan(
		&user.ID, &user.Username, &user.Email, &user.IsStaff, &user.IsSuperuser)
	return user, err
}
