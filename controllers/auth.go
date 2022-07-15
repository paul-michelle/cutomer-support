package controllers

import (
	"db-queries/db"
	"encoding/json"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JWT_KEY = []byte(os.Getenv("JWT_KEY"))

const COOKIE_NAME = "token"
const TOKEN_TTL_MINS = 5

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Claims struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	IsStaff     bool   `json:"isStaff"`
	IsSuperuser bool   `json:"isSuperuser"`
	jwt.RegisteredClaims
}

func (h *BaseHandler) LogIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Email and password required to obtain a token.", http.StatusBadRequest)
		return
	}

	_, emailParseError := mail.ParseAddress(creds.Email)
	if emailParseError != nil || creds.Password == "" {
		http.Error(w, "Valid email address and password required.", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserDetails(h.Conn, creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "User with specified credentials not found.", http.StatusNotFound)
		return
	}

	ttl := time.Now().Add(TOKEN_TTL_MINS * time.Minute)
	tokenString, err := createTokenForUser(user, ttl)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    COOKIE_NAME,
		Value:   tokenString,
		Expires: ttl,
	})
}

func createTokenForUser(user db.User, ttl time.Time) (tokenString string, err error) {
	claims := &Claims{
		Username:    user.Username,
		Email:       user.Email,
		IsStaff:     user.IsStaff,
		IsSuperuser: user.IsSuperuser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(ttl),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_KEY)
}
