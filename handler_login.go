package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Relevantfender/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	type reqVals struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type respVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}
	// seconds in an hour * 100 for nano seconds
	oneHour := 3600
	request := reqVals{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&request)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while unmarshaling the request for login", err)
		return
	}

	if request.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide the email", err)
		return
	}

	if request.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide the password", err)
		return
	}

	if request.ExpiresInSeconds == 0 || request.ExpiresInSeconds > 3600 {
		request.ExpiresInSeconds = oneHour
	}

	user, err := cfg.dbQueries.FindUserByEmail(r.Context(), request.Email)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error while retrieving user from database", err)
		return
	}

	err = auth.CheckPasswordHash(request.Password, user.HashedPassword)
	if err != nil {
		respondWithError(
			w,
			http.StatusUnauthorized,
			"Wrong password",
			err,
		)
		return
	}

	jwtToken, err := auth.MakeJWT(
		user.ID,
		cfg.jwtSecret,
		time.Duration(request.ExpiresInSeconds)*time.Second,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while creating a jwt token", err)
	}
	response := respVals{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     jwtToken,
	}

	respondWithJSON(
		w,
		http.StatusOK,
		response,
	)

}
