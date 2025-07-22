package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Relevantfender/internal/auth"
	"github.com/Relevantfender/internal/database"
	"github.com/google/uuid"
)

type request struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
type response struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {

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
		respondWithError(w, http.StatusInternalServerError, "error while creating the user in handlerCreateUser, line 57", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, response{
		ID:          userData.ID,
		CreatedAt:   userData.CreatedAt,
		UpdatedAt:   userData.UpdatedAt,
		Email:       userData.Email,
		IsChirpyRed: userData.IsChirpyRed,
	})

}

func (cfg *apiConfig) handleUpdateCredentials(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		log.Printf("Error while getting the token  out of the header: %v", err)
		return

	}

	request := request{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		log.Printf("Error while decoding the request in update credentials handler:%v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while decoding the request in update credentials handler", err)
		return
	}

	if request.Email == "" {
		log.Print("email is empty in handler update credentials request")
		respondWithError(w, http.StatusBadRequest, "Provide the email in request", err)
		return
	}
	if request.Email == "" {
		log.Print("password is empty in handler update credentials request")
		respondWithError(w, http.StatusBadRequest, "Provide the password in request", err)
		return

	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		log.Printf("Error while validating the jwt token:%v", err)
		respondWithError(w, http.StatusUnauthorized, "Error while getting user from db", err)
		return

	}

	_, err = cfg.dbQueries.FindUserByID(r.Context(), userID)

	if err != nil {
		log.Printf("Error while getting user from db:%v", err)
		respondWithError(w, http.StatusNotFound, "Error while getting user from db", err)
		return
	}

	hashedPassword, err := auth.HashPassword(request.Password)

	if err != nil {
		log.Printf("Error while hashing the password:%v", err)
		respondWithError(w, http.StatusNotFound, "Error while hashing the password", err)
		return
	}

	userParams := database.UpdatedUserParams{
		Email:          request.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}

	user, err := cfg.dbQueries.UpdatedUser(r.Context(), userParams)

	if err != nil {
		log.Printf("Error while updating the user:%v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while updating the user", err)
		return
	}

	response := response{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusOK, response)
}
