package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type jsonValidateResp struct {
	Valid       bool   `json:"valid,omitempty"`
	Error       string `json:"error,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
}

type newChirp struct {
	Body string `json:"Body"`
}

var Profanity []string

func (c *apiConfig) handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	Profanity = append(Profanity, "kerfuffle", "sharbert", "fornax")
	w.Header().Add("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	var jsonResponse jsonValidateResp
	var chirp newChirp
	err := decoder.Decode(&chirp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Valid = false
		jsonResponse.Error = fmt.Sprintf("Something went wrong: %v", err)
	} else if chirp.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Valid = false
		jsonResponse.Error = "Chirp is not present"
	} else if len(chirp.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Valid = false
		jsonResponse.Error = "Chirp is too long"
	} else {
		w.WriteHeader(http.StatusOK)
		jsonResponse.Valid = true
		words := strings.Split(chirp.Body, " ")
		var clean []string
		for _, word := range words {
			dirty := slices.Contains(Profanity, strings.ToLower(word))
			if dirty {
				clean = append(clean, "****")
			} else {
				clean = append(clean, word)
			}
		}
		cleaned := strings.Join(clean, " ")
		jsonResponse.CleanedBody = cleaned
	}
	jsonOut, err := json.Marshal(jsonResponse)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Valid = false
		jsonResponse.Error = fmt.Sprintf("Something went wrong: %v", err)
	}
	w.Write(jsonOut)
}
