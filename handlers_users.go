package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/auth"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Email     string    `json:"email,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type userParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *apiConfig) handlerUserRegistration(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var jsonResponse User
	decoder := json.NewDecoder(req.Body)
	var newUser userParams
	err := decoder.Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Error = fmt.Sprintf("error decoding request: %v", err)
	} else {
		hashedPW, err := auth.HashPassword(newUser.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonResponse.Error = fmt.Sprintf("error creating password hash: %v", err)
		} else {
			newUserParams := database.CreateUserParams{
				Email:          newUser.Email,
				HashedPassword: hashedPW,
			}
			userResp, err := c.dbQueries.CreateUser(req.Context(), newUserParams)
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
	}

	jsonOut, err := json.Marshal(jsonResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse.Error = fmt.Sprintf("Something went wrong compiling output: %v", err)
	}
	w.Write(jsonOut)
}

func (c *apiConfig) handlerUserLogin(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	var jsonResponse User
	var user userParams
	err := decoder.Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse.Error = fmt.Sprintf("error decoding request: %v", err)
	} else {
		userResp, err := c.dbQueries.GetUserPWHashByEmail(req.Context(), user.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonResponse.Error = fmt.Sprintf("error reteving PW hash: %v", err)
		} else {
			pwMatch, err := auth.CheckPasswordHash(user.Password, userResp.HashedPassword)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				jsonResponse.Error = fmt.Sprintf("error comparing PW with hash: %v", err)
			} else if pwMatch {
				w.WriteHeader(http.StatusOK)
				jsonResponse.ID = userResp.ID
				jsonResponse.CreatedAt = userResp.CreatedAt
				jsonResponse.UpdatedAt = userResp.UpdatedAt
				jsonResponse.Email = userResp.Email
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				jsonResponse.Email = "Email or password missmatch"
			}
		}
	}

	jsonOut, err := json.Marshal(jsonResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse.Error = fmt.Sprintf("Something went wrong compiling output: %v", err)
	}
	w.Write(jsonOut)
}
