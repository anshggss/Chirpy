package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenString string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	s, err := token.SignedString([]byte(tokenString))
	if err != nil {
		return "", err
	}
	return s, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {

		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("Signing method of the token doesn't match")
		}
		return []byte(tokenSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, err
	}
	s, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	uid, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil

}
func GetBearerToken(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return val, fmt.Errorf("No Authorization header/token found")
	}
	return strings.Fields(val)[1], nil

}

func MakeRefreshToken() string {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	return hex.EncodeToString(randBytes)
}
