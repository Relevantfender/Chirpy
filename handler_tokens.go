package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Relevantfender/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error while obtaining refresh token", err)
		return
	}

	dbRefreshToken, err := cfg.dbQueries.FindRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error while finding the refresh token in db: %v", err)
		respondWithError(w, http.StatusUnauthorized, "No refresh token found", err)
		return
	}

	// Valid is false if Time is NULL
	if dbRefreshToken.RevokedAt.Valid {
		log.Print("refresh token has been revoked")
		respondWithError(w, http.StatusUnauthorized, "token has been revoked", nil)
		return
	}

	if dbRefreshToken.ExpiresAt.Time.Before(time.Now()) {
		log.Print("refresh token has expired")
		respondWithError(w, http.StatusUnauthorized, "refresh token has expired", nil)
		return
	}

	userID := dbRefreshToken.UserID

	user, err := cfg.dbQueries.FindUserByID(r.Context(), userID)

	if err != nil {
		log.Print("No user found")
		respondWithError(w, http.StatusNotFound, "db error, no user found", err)
		return
	}

	newAccessToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Hour,
	)
	if err != nil {
		log.Print("Error creating a new access token")
		respondWithError(w, http.StatusInternalServerError, "error creating a new access token", err)
		return
	}

	respValue := struct {
		Token string `json:"token"`
	}{
		Token: newAccessToken,
	}

	respondWithJSON(w, http.StatusOK, respValue)
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting the bearer token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "error extracting the token, line 72", err)
		return
	}

	err = cfg.dbQueries.RevokeToken(r.Context(), refreshToken)

	if err != nil {
		log.Print("Error creating a new access token")
		respondWithError(w, http.StatusInternalServerError, "error creating a new access token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
