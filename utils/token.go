package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type UserIDKey struct{}

// TODO modify to support various token types
type TokenCustomClaims struct {
	Username string
	jwt.StandardClaims
}

func GenerateToken(userName string) (string, error) {
	claims := TokenCustomClaims{
		userName,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret, err := AccessSecretVersion(SECRET_USER_KEY)
	if err != nil {
		return "", err
	}
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func WithToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authString := r.Header.Get("Authorization")
		if authString == "" {
			RespondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}
		splitAuth := strings.Fields(authString)
		u, err := validateToken(splitAuth[1])
		if err != nil {
			RespondWithError(w, 400, err.Error())
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey{}, u)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func validateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		secret, err := AccessSecretVersion(SECRET_USER_KEY)
		if err != nil {
			return "", err
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*TokenCustomClaims)
	if !ok || !token.Valid || claims.Username == "" {
		return "", fmt.Errorf("invalid token")
	}

	return claims.Username, nil
}
