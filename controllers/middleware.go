package controllers

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

var wrongSigningMethodError = errors.New("Unexpected signing method")

func JWTMiddleWare(next func(res http.ResponseWriter, req *AuthenticatedRequest)) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		tokenString, errorNoCookie := req.Cookie(COOKIE_NAME)
		if errorNoCookie != nil {
			http.Error(res, "Token missing in 'Cookie' headers.", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		_, validErr := jwt.ParseWithClaims(tokenString.Value, claims,
			func(tkn *jwt.Token) (interface{}, error) {
				if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, wrongSigningMethodError
				}
				return JWT_KEY, nil
			})

		if validErr != nil {
			msg := "Token invalid."
			errParsed, _ := validErr.(*jwt.ValidationError)
			if errParsed.Errors == jwt.ValidationErrorExpired {
				msg = "Token expired."
			}
			http.Error(res, msg, http.StatusUnauthorized)
			return
		}

		enrichedReruest := &AuthenticatedRequest{req, Requester{
			IsStaff:     claims.IsStaff,
			IsSuperuser: claims.IsSuperuser,
			Email:       claims.Email,
			Username:    claims.Username,
		},
		}
		next(res, enrichedReruest)
	})
}
