package db

import "database/sql"

const (
	createUserStmt = `
	INSERT INTO users (email, password, username, is_staff, is_superuser) 
	VALUES ($1, crypt($2, gen_salt('bf', 8)), $3, $4, $5);`

	getUserDetailsStmt = `
	SELECT id, username, email, is_staff, is_superuser FROM  users 
	WHERE email=$1 and password=crypt($2, password);`
)

type User struct {
	ID          int
	Username    string
	Email       string
	IsStaff     bool
	IsSuperuser bool
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
