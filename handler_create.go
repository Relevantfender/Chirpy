package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Relevantfender/internal/auth"
	"github.com/Relevantfender/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	req := request{}
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error while decoding the json", err)
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide the email", nil)
		return
	}
	if req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide the password", nil)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while hashing the password", err)
		return
	}

	dto := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hash,
	}
	userData, err := cfg.dbQueries.CreateUser(r.Context(), dto)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error while creating the user", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, response{
		ID:        userData.ID,
		CreatedAt: userData.CreatedAt,
		UpdatedAt: userData.UpdatedAt,
		Email:     userData.Email,
	})

}
