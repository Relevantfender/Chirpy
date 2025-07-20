package auth

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	t.Run("successful token creation", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if token == "" {
			t.Fatal("Expected non-empty token")
		}

		// Verify the token can be parsed and contains expected claims
		claims := &jwt.RegisteredClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})
		if err != nil {
			t.Fatalf("Failed to parse generated token: %v", err)
		}
		if !parsedToken.Valid {
			t.Fatal("Generated token is not valid")
		}
		if claims.Issuer != "chirpy" {
			t.Errorf("Expected issuer 'chirpy', got %s", claims.Issuer)
		}
		if claims.Subject != userID.String() {
			t.Errorf("Expected subject %s, got %s", userID.String(), claims.Subject)
		}
		if claims.IssuedAt == nil {
			t.Error("Expected IssuedAt to be set")
		}
		if claims.ExpiresAt == nil {
			t.Error("Expected ExpiresAt to be set")
		}
	})

	t.Run("different expiration times", func(t *testing.T) {
		shortExpiry := time.Minute
		longExpiry := time.Hour * 24

		shortToken, err := MakeJWT(userID, tokenSecret, shortExpiry)
		if err != nil {
			t.Fatalf("Expected no error for short expiry, got %v", err)
		}

		longToken, err := MakeJWT(userID, tokenSecret, longExpiry)
		if err != nil {
			t.Fatalf("Expected no error for long expiry, got %v", err)
		}

		if shortToken == longToken {
			t.Error("Expected different tokens for different expiry times")
		}
	})

	t.Run("different user IDs produce different tokens", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()

		token1, err := MakeJWT(userID1, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Expected no error for user1, got %v", err)
		}

		token2, err := MakeJWT(userID2, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Expected no error for user2, got %v", err)
		}

		if token1 == token2 {
			t.Error("Expected different tokens for different user IDs")
		}
	})

	t.Run("different secrets produce different tokens", func(t *testing.T) {
		secret1 := "secret1"
		secret2 := "secret2"

		token1, err := MakeJWT(userID, secret1, expiresIn)
		if err != nil {
			t.Fatalf("Expected no error for secret1, got %v", err)
		}

		token2, err := MakeJWT(userID, secret2, expiresIn)
		if err != nil {
			t.Fatalf("Expected no error for secret2, got %v", err)
		}

		if token1 == token2 {
			t.Error("Expected different tokens for different secrets")
		}
	})

	t.Run("zero expiration time", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, 0)
		if err != nil {
			t.Fatalf("Expected no error for zero expiration, got %v", err)
		}
		if token == "" {
			t.Fatal("Expected non-empty token even with zero expiration")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		token, err := MakeJWT(userID, "", expiresIn)
		if err != nil {
			t.Fatalf("Expected no error for empty secret, got %v", err)
		}
		if token == "" {
			t.Fatal("Expected non-empty token even with empty secret")
		}
	})
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	t.Run("successful token validation", func(t *testing.T) {
		// Create a valid token first
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Validate the token
		validatedUserID, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if validatedUserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("invalid token string", func(t *testing.T) {
		invalidToken := "invalid.token.string"
		_, err := ValidateJWT(invalidToken, tokenSecret)
		if err == nil {
			t.Error("Expected error for invalid token string")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		// Create token with one secret
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Try to validate with different secret
		wrongSecret := "wrong-secret"
		_, err = ValidateJWT(token, wrongSecret)
		if err == nil {
			t.Error("Expected error for wrong secret")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		// Create token that expires immediately
		expiredToken, err := MakeJWT(userID, tokenSecret, -time.Hour)
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}

		// Try to validate expired token
		_, err = ValidateJWT(expiredToken, tokenSecret)
		if err == nil {
			t.Error("Expected error for expired token")
		}
	})

	t.Run("empty token string", func(t *testing.T) {
		_, err := ValidateJWT("", tokenSecret)
		if err == nil {
			t.Error("Expected error for empty token string")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		// Create token with empty secret
		token, err := MakeJWT(userID, "", expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Validate with empty secret
		validatedUserID, err := ValidateJWT(token, "")
		if err != nil {
			t.Fatalf("Expected no error for empty secret, got %v", err)
		}
		if validatedUserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("token with invalid signing method", func(t *testing.T) {
		// Create a token header that claims to use RS256
		header := `{"alg":"RS256","typ":"JWT"}`
		payload := `{"iss":"chirpy","sub":"` + userID.String() + `","iat":1516239022}`

		// Create a malformed token (base64 encoded header.payload.invalid_signature)
		encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(header))
		encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
		malformedToken := encodedHeader + "." + encodedPayload + ".invalid_signature"

		_, err := ValidateJWT(malformedToken, tokenSecret)
		if err == nil {
			t.Error("Expected error for token with wrong signing method")
		}
	})

	t.Run("token without subject", func(t *testing.T) {
		// Create token without subject
		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			// Subject is omitted
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(tokenSecret))
		if err != nil {
			t.Fatalf("Failed to create token without subject: %v", err)
		}

		_, err = ValidateJWT(tokenString, tokenSecret)
		if err == nil {
			t.Error("Expected error for token without subject")
		}
	})

	t.Run("token with invalid UUID in subject", func(t *testing.T) {
		// Create token with invalid UUID as subject
		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   "invalid-uuid",
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(tokenSecret))
		if err != nil {
			t.Fatalf("Failed to create token with invalid UUID: %v", err)
		}

		_, err = ValidateJWT(tokenString, tokenSecret)
		if err == nil {
			t.Error("Expected error for token with invalid UUID in subject")
		}
	})
}

func TestMakeJWTAndValidateJWT_Integration(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "integration-test-secret"
	expiresIn := time.Hour

	t.Run("round trip test", func(t *testing.T) {
		// Create token
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Validate token
		validatedUserID, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		// Check if user ID matches
		if validatedUserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("multiple users round trip", func(t *testing.T) {
		userIDs := []uuid.UUID{
			uuid.New(),
			uuid.New(),
			uuid.New(),
		}

		for i, uid := range userIDs {
			token, err := MakeJWT(uid, tokenSecret, expiresIn)
			if err != nil {
				t.Fatalf("Failed to create token for user %d: %v", i, err)
			}

			validatedUserID, err := ValidateJWT(token, tokenSecret)
			if err != nil {
				t.Fatalf("Failed to validate token for user %d: %v", i, err)
			}

			if validatedUserID != uid {
				t.Errorf("User %d: Expected userID %s, got %s", i, uid, validatedUserID)
			}
		}
	})
}
