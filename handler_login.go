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
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type respVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

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

	user, err := cfg.dbQueries.FindUserByEmail(r.Context(), request.Email)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error while retrieving user from database", err)
		return
	}

	// hashedPass, err := auth.HashPassword(request.Password)
	// if err != nil {
	// 	respondWithError(
	// 		w,
	// 		http.StatusInternalServerError,
	// 		"Error while hashing the password from request while login",
	// 		err,
	// 	)
	// }

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

	response := respVals{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(
		w,
		http.StatusOK,
		response,
	)

}
