package controllers

import (
	"database/sql"
	"db-queries/db"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lib/pq"
)

const (
	passwordMinLength                   = 8
	tokenTtlMinutes                     = 5
	maxMinsBeforeExpTokenCanBeRefreshed = 1
	cookieName                          = "token"
)

var (
	wrongSigningMethodError = errors.New("Unexpected signing method")
	jwtKey                  = []byte(os.Getenv("JWT_KEY"))
)

type BaseHandler struct {
	Conn *sql.DB
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{db}
}

type TicketDetails struct {
	Author   int    `json:"author"`
	Topic    string `json:"topic"`
	Contents string `json:"contents"`
}

type UserDetails struct {
	IsStaff     bool   `json:"isStaff"`
	IsSuperuser bool   `json:"isSuperuser"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username"`
}

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

func (h *BaseHandler) Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}

func (h *BaseHandler) TicketsListAllOrCreateOne(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.GetAllTickets(w, r)
		return
	}
	if r.Method == "POST" {
		h.CreateTicket(w, r)
		return
	}
	http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
}

func (h *BaseHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Payload expected.", http.StatusBadRequest)
		return
	}

	var ticket TicketDetails
	err := json.NewDecoder(r.Body).Decode(&ticket)
	if err != nil {
		http.Error(w, "Invalid payload.", http.StatusBadRequest)
		return
	}

	if ticket.Author == 0 || ticket.Topic == "" || ticket.Contents == "" {
		http.Error(w, "Missing fields in payload", http.StatusBadRequest)
		return
	}

	id, err := db.CreateTicket(h.Conn, ticket.Author, ticket.Topic)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := make(map[string]int)
	resp["id"] = id
	json.NewEncoder(w).Encode(resp)
}

func (h *BaseHandler) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := db.GetAllTickets(h.Conn)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets.Tickets)
}

func (h *BaseHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	var user UserDetails
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid payload.", http.StatusBadRequest)
		return
	}

	_, emailParseError := mail.ParseAddress(user.Email)
	if emailParseError != nil || user.Password == "" || user.Username == "" {
		http.Error(w, "Username, password and valid email address required.", http.StatusBadRequest)
		return
	}

	if len(user.Password) < passwordMinLength {
		http.Error(w, fmt.Sprintf("Password min length is %d", passwordMinLength), http.StatusBadRequest)
		return
	}

	staffStatus := false
	if user.IsStaff {
		reqToken := r.Header.Get("Authorization")
		if reqToken == "" {
			http.Error(w, "Token missing.", http.StatusUnauthorized)
			return
		}

		splitToken := strings.Split(reqToken, "Bearer")
		if len(splitToken) != 2 {
			http.Error(w, "Token of wrong format.", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(splitToken[1])
		if token != os.Getenv("STAFF_TOKEN") {
			http.Error(w, "Token invalid", http.StatusUnauthorized)
			return
		}

		staffStatus = true
	}

	if err := db.CreateUser(h.Conn, user.Email, user.Password, user.Username, staffStatus, false); err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code.Name() == "unique_violation" {
			http.Error(w, "User with specified email already exists.", http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
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

	ttl := time.Now().Add(tokenTtlMinutes * time.Minute)
	tokenString, err := createTokenForUser(user, ttl)
	if err != nil {
		http.Error(w, "Please try again later.", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    cookieName,
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
	return token.SignedString(jwtKey)
}
