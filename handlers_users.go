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
	ID           uuid.UUID `json:"id,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	Email        string    `json:"email,omitempty"`
	Error        string    `json:"error,omitempty"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type Token struct {
	Error string `json:"error,omitempty"`
	Token string `json:"token,omitempty"`
}

type userParams struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	ExpiresIn int64  `json:"expires_in_seconds"`
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
	var newRefreshTokenParams database.RefreshTokenNewParams
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
				token, err := auth.MakeJWT(userResp.ID, c.tokenSecret, c.tokenDefualtDuration)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					jsonResponse.Error = fmt.Sprintf("error creating token: %v", err)
				} else {
					refreshToken, err := auth.MakeRefeshToken()
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						jsonResponse.Error = fmt.Sprintf("error creating refresh token: %v", err)
					} else {
						newRefreshTokenParams.Token = refreshToken
						newRefreshTokenParams.UserID = userResp.ID
						newRefreshTokenParams.ExpiresAt = (time.Now()).Add(c.refreshTokenTimeout)

						newRefreshTokenReturn, err := c.dbQueries.RefreshTokenNew(req.Context(), newRefreshTokenParams)
						if err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							jsonResponse.Error = fmt.Sprintf("error creating refresh token: %v", err)
						} else {
							w.WriteHeader(http.StatusOK)
							jsonResponse.ID = userResp.ID
							jsonResponse.CreatedAt = userResp.CreatedAt
							jsonResponse.UpdatedAt = userResp.UpdatedAt
							jsonResponse.Email = userResp.Email
							jsonResponse.Token = token
							jsonResponse.RefreshToken = newRefreshTokenReturn.Token
						}
					}
				}
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

func (c *apiConfig) handlerRefreshToken(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var jsonResponse Token

	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		jsonResponse.Error = fmt.Sprintf("error creating refresh token: %v", err)
	} else {
		refreshTokenUser, err := c.dbQueries.RefreshTokenGet(req.Context(), refreshToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonResponse.Error = fmt.Sprintf("error checking refresh token: %v", err)
		} else if refreshTokenUser.RevokedAt.Valid == false {
			newJWT, err := auth.MakeJWT(refreshTokenUser.UserID, c.tokenSecret, c.tokenDefualtDuration)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				jsonResponse.Error = fmt.Sprintf("error creating new token: %v", err)
			} else {
				w.WriteHeader(http.StatusOK)
				jsonResponse.Token = newJWT
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			jsonResponse.Error = "Refresh token has been revoked"
		}
	}

	jsonOut, err := json.Marshal(jsonResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse.Error = fmt.Sprintf("Something went wrong compiling output: %v", err)
	}
	w.Write(jsonOut)
}

func (c *apiConfig) handlerRefreshTokenRevoke(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprintf("Refesh token not found: %v", err)))
	} else {
		_, err := c.dbQueries.RefreshTokenRevoke(req.Context(), refreshToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error revoking refresh token: %v", err)))
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
