package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	clock := time.Now().UTC()
	currentTime := jwt.NewNumericDate(clock)
	expiryDate := jwt.NewNumericDate(clock.Add(expiresIn))
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  currentTime,
		ExpiresAt: expiryDate,
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	jwtToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		// func that checks the header for the singing method hmac, and if ok it returns the
		// token secret as an array of bytes
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
	)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !jwtToken.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	if claims.Subject == "" {
		return uuid.Nil, fmt.Errorf("no subject found in token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID in subject: %w", err)
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	bearerVal := strings.Split(bearer, "Bearer ")
	if len(bearerVal) < 2 {
		return "", errors.New("no bearer token attached")
	}

	// if err!=nil{
	// 	fmt.Errorf("Error while getting bearer token in jwt.go, line 75: %w", err)
	// }

	return bearerVal[1], nil
}
