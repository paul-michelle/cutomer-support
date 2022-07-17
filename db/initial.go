package db

import (
	"database/sql"
	"fmt"
	"log"

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
				CREATE TYPE status AS ENUM ('resolved', 'unresolved', 'pending', 'canceled');
			END IF;
			RETURN type_already_exists;
		END;
		$$ LANGUAGE plpgsql;
	SELECT create_ticket_status_type();
	DROP function create_ticket_status_type();`

	CREATE_TABLE_TICKETS_STMT = `
	CREATE TABLE IF NOT EXISTS tickets
	(
		id SERIAL,
		created_at TIMESTAMP DEFAULT now(),
		updated_at TIMESTAMP DEFAULT now(),
		author VARCHAR(64) REFERENCES users (email) ON DELETE CASCADE,
		topic VARCHAR(20) NOT NULL,
		status STATUS,
		CONSTRAINT pk_tickets PRIMARY KEY (id)
	);`

	createTableMessagesStmt = `
	CREATE TABLE IF NOT EXISTS messages
	(
		id SERIAL,
		created_at TIMESTAMP DEFAULT now(),
		author VARCHAR(64) REFERENCES users (email) ON DELETE CASCADE,
		text TEXT,
		ticket INTEGER REFERENCES tickets (id) ON DELETE CASCADE,
		CONSTRAINT pk_messages PRIMARY KEY (id)
	);`
	VALUE_TOO_LONG_ERR_CODE_NAME   = "string_data_right_truncation"
	UNIQUE_VIOLATION_ERR_CODE_NAME = "unique_violation"
)

var (
	VALID_STATUSES                  = []string{"resolved", "unresolved", "canceled", "pending"}
	VALID_TICKET_STATUS_COMMON_USER = map[string]bool{
		"canceled": true,
	}
	VALID_TICKET_STATUS_STAFF = map[string]bool{
		"resolved":   true,
		"unresolved": true,
		"pending":    true,
	}
)

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
	_, err = conn.Exec(CREATE_TABLE_TICKETS_STMT)
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
