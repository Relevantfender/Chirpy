package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Relevantfender/internal/auth"
	"github.com/Relevantfender/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	type reqVals struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type respVals struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}
	// seconds in an hour
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
		time.Duration(oneHour)*time.Second,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while creating a jwt token", err)
	}

	refToken, _ := auth.MakeRefreshToken()

	refTokenVals := database.CreateRefreshTokenParams{
		Token:  refToken,
		UserID: user.ID,
		ExpiresAt: sql.NullTime{
			Time:  time.Now().Add(60 * 24 * time.Hour),
			Valid: true,
		},
	}

	rToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), refTokenVals)
	if err != nil {
		fmt.Printf("Error during saving of the refresh token in db: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while saving refresh token in db", err)
		return
	}
	response := respVals{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwtToken,
		RefreshToken: rToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(
		w,
		http.StatusOK,
		response,
	)

}

func (cfg *apiConfig) handleUpdatePolkaUpgrade(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserId string `json:"user_id"`
	}
	type reqVals struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	request := reqVals{}

	decoder.Decode(&request)

	apiKey, err := auth.GetAPIKey(r.Header)

	if err != nil {
		fmt.Printf("Error while getting the api key from the header: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Error, wrong api key", err)
		return
	}

	if apiKey != cfg.polkaKey {
		fmt.Print("Wrong api key from polka")
		respondWithError(w, http.StatusUnauthorized, "Error, wrong api key", err)
		return
	}

	if request.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, "The message is not user.upgraded")
		return
	}
	userID, err := uuid.Parse(request.Data.UserId)

	if err != nil {
		fmt.Printf("Error while parsing the ID from the request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Error while parsing the ID from the request body", err)
		return
	}

	dto := database.UpdateUserChirpyRedParams{
		IsChirpyRed: true,
		ID:          userID,
	}
	err = cfg.dbQueries.UpdateUserChirpyRed(r.Context(), dto)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("Error while Updating the user chirpy red: %v", err)
			respondWithError(w, http.StatusNotFound, "No user has been found with that id", err)
			return
		}
		fmt.Printf("Error while Updating the user chirpy red: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error while Updating the user chirpy red", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "The user has been upgraded")

}
