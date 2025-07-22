package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/Relevantfender/internal/auth"
	"github.com/Relevantfender/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {

	type requestVals struct {
		Body    string    `json:"body"`
		User_id uuid.UUID `json:"user_id"`
	}

	type responseVals struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		User_id    uuid.UUID `json:"user_id"`
	}
	request := requestVals{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&request)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Error while decoding",
			err,
		)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error while retrieving bearer:", err)
		return
	}

	tokenId, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		log.Printf("Error while validating token: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Error while validating token:", err)
		return
	}

	if request.Body == "" {
		respondWithError(
			w,
			http.StatusBadRequest,
			"Please provide the chirp",
			nil,
		)
		return
	}
	if (tokenId == uuid.UUID{}) {
		respondWithError(
			w,
			http.StatusBadRequest,
			"Please provide the user id",
			nil,
		)
		return

	}

	const maxChirpLength = 140
	if len(request.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(request.Body, badWords)

	request.Body = cleaned
	args := database.CreateChirpParams{
		Body:   cleaned,
		UserID: tokenId,
	}
	response, err := cfg.dbQueries.CreateChirp(r.Context(), args)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while creating the DB entry", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, responseVals{
		Id:         response.ID,
		Created_at: response.CreatedAt,
		Updated_at: response.UpdatedAt,
		Body:       response.Body,
		User_id:    response.UserID,
	})

}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {

	type responseVals struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		UserId     uuid.UUID `json:"user_id"`
	}
	authorId := r.URL.Query().Get("author_id")
	sortingOrder := r.URL.Query().Get("sort")
	if sortingOrder == "" {
		sortingOrder = "asc"
	}

	var chirps []database.Chirp
	var err error

	if authorId != "" {
		uuid, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Invalid author id", err)
			return
		}
		chirps, err = cfg.dbQueries.GetChirpsByAuthorId(r.Context(), uuid)
	} else {
		chirps, err = cfg.dbQueries.GetChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusNotFound, "No chirps found", err)
		return
	}

	if sortingOrder == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[j].CreatedAt.Before(chirps[i].CreatedAt)
		})
	}
	var response []responseVals

	for _, chirp := range chirps {
		response = append(response, responseVals{
			Id:         chirp.ID,
			Created_at: chirp.CreatedAt,
			Updated_at: chirp.UpdatedAt,
			Body:       chirp.Body,
			UserId:     chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, response)

}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	type responseVals struct {
		Id         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Body       string    `json:"body"`
		UserID     uuid.UUID `json:"user_id"`
	}
	id := r.PathValue("chirpID")
	uid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while parsing the uuid", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpsById(r.Context(), uid)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "chirp not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while parsing the uuid", err)
		return
	}
	response := responseVals{
		Id:         chirp.ID,
		Created_at: chirp.CreatedAt,
		Updated_at: chirp.UpdatedAt,
		Body:       chirp.Body,
		UserID:     chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	path_value := r.PathValue("chirpID")
	if path_value == "" {
		respondWithError(w, http.StatusBadRequest, "Please provide a chirp id", nil)
		log.Print("No chirp id in path")
		return
	}

	chirpID, err := uuid.Parse(path_value)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad id value passed", err)
		log.Print("Bad id value passed")
		return
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error in getting bearer token", err)
		log.Printf("Error while getting bearer token: %v", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error in validating the token", err)
		log.Printf("Error while validating a jwt: %v", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpsById(r.Context(), chirpID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "No chirps found", err)
			log.Printf("Error while validating a jwt: %v", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error while querying the chirps", err)
		log.Printf("Error while querying the chirps: %v", err)
		return
	}
	if userId != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Not allowed", err)
		log.Printf("Error while checking the id of the author: %v", err)
		return
	}

	err = cfg.dbQueries.DeleteChirpById(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while deleting the resource chirp", err)
		log.Printf("Error while deleting the resource chirp: %v", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)

}
