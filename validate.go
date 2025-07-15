package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json: "body"`
	}
	type returnVals struct {
		Valid bool   `json:"valid"`
		Error string `json: "error"`
	}
	decoder := json.NewDecoder(r.Body)
	requestBody := parameters{}
	err := decoder.Decode(&requestBody)
	if err != nil {
		log.Printf("Error while decoding body parameters: %s", err)
		response := returnVals{
			Valid: false,
			Error: "Something went wrong"}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error while marshaling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprint("Error while Marshaling")))
			return
		}
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	if requestBody.Body == "" {
		log.Printf("Bad request, no body present")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide a body"))
		return
	}

	if len(requestBody.Body) > 150 {
		w.WriteHeader(http.StatusBadRequest)
		response := returnVals{
			Valid: false,
			Error: "Chirp is too long"}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error while marshaling json object at validate line 40: %s", err)
			w.Write([]byte("Error while marhsaling json object"))
			return
		}
		w.Write(data)
		return
	}

	vals := strings.Split(requestBody.Body, " ")
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	for i := 0; i < len(vals); i++ {
		for j := 0; j < len(profaneWords); j++ {
			if vals[i] == profaneWords[j] {
				vals[i] = "****"
			}

		}
	}

	requestBody.Body = strings.Join(vals, " ")

	response := returnVals{Valid: true}
	data, err := json.Marshal(response)

	if err != nil {
		log.Printf("Error while Marhsalling line 59")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while Marshaling json object"))
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
