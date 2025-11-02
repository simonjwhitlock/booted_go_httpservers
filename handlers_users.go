package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Email     string    `json:"email,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type newUserEmail struct {
	Email string `json:"email"`
}

func (c *apiConfig) handlerUserRegistration(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	var jsonResponse User
	var newUser newUserEmail
	err := decoder.Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Error = fmt.Sprintf("error decoding request: %v", err)
	} else {
		userResp, err := c.dbQueries.CreateUser(req.Context(), newUser.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonResponse.Error = fmt.Sprintf("error create user: %v", err)
		} else {
			w.WriteHeader(http.StatusCreated)
			jsonResponse.ID = userResp.ID
			jsonResponse.CreatedAt = userResp.CreatedAt
			jsonResponse.UpdatedAt = userResp.UpdatedAt
			jsonResponse.Email = userResp.Email
		}

	}

	jsonOut, err := json.Marshal(jsonResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse.Error = fmt.Sprintf("Something went wrong compiling output: %v", err)
	}
	w.Write(jsonOut)
}
