package db

import (
	"database/sql"
	"time"
)

const (
	createUserStmt = `
	INSERT INTO users (email, password, username, is_staff, is_superuser) 
	VALUES ($1, crypt($2, gen_salt('bf', 8)), $3, $4, $5);`

	getUserDetailsStmt = `
	SELECT id, username, email, is_staff, is_superuser FROM  users 
	WHERE email=$1 and password=crypt($2, password);`

	GET_ALL_USERS = `
	SELECT u.id, u.created_at, u.username, u.email, u.is_staff, count(t.id) as ticketsCount
	FROM users u LEFT JOIN tickets t ON u.email = t.author
	GROUP BY u.id
	ORDER BY ticketsCount DESC
	`
)

type User struct {
	ID           int       `json:"id"`
	CrtdAt       time.Time `json:"created_at"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	IsStaff      bool      `json:"is_staff"`
	IsSuperuser  bool      `json:"is_superuser,omitempty"`
	TicketsCount int       `json:"tickets_count"`
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

func GetAllUsers(conn *sql.DB) ([]User, error) {
	rows, err := conn.Query(GET_ALL_USERS)
	if err != nil {
		return nil, err
	}

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.CrtdAt, &u.Username, &u.Email, &u.IsStaff, &u.TicketsCount)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
