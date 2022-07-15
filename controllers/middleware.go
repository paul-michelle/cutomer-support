package controllers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

func JWTMiddleWare(next func(res http.ResponseWriter, req *http.Request)) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		tokenString, errorNoCookie := req.Cookie(cookieName)
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
				return jwtKey, nil
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

		next(res, req)
	})
}
