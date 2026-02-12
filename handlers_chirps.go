package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/auth"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/database"
)

type jsonValidateResp struct {
	Error     string    `json:"error,omitempty"`
	Body      string    `json:"body,omitempty"`
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
}

type newChirp struct {
	Body   string    `json:"Body"`
	UserID uuid.UUID `json:"user_id"`
}

var Profanity []string

func (c *apiConfig) handlerChirps(w http.ResponseWriter, req *http.Request) {
	Profanity = append(Profanity, "kerfuffle", "sharbert", "fornax")

	userID, err := auth.TokenAuth(req.Header, c.tokenSecret)
	if err != nil {
		fmt.Printf("error with header auth token: %v", err)
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
	} else {
		w.Header().Add("Content-Type", "application/json")

		decoder := json.NewDecoder(req.Body)
		var jsonResponse jsonValidateResp
		var chirp newChirp
		err = decoder.Decode(&chirp)
		chirp.UserID = userID
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			jsonResponse.Error = fmt.Sprintf("Something went wrong: %v", err)
		} else if chirp.Body == "" {
			w.WriteHeader(http.StatusBadRequest)
			jsonResponse.Error = "Chirp is not present"
		} else if len(chirp.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			jsonResponse.Error = "Chirp is too long"
		} else {
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
			cleanedChirp := database.NewChirpParams{
				Body:   strings.Join(clean, " "),
				UserID: chirp.UserID,
			}
			chirpResp, err := c.dbQueries.NewChirp(req.Context(), cleanedChirp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				jsonResponse.Error = fmt.Sprintf("error submitting chirp to db: %v", err)
			} else {
				w.WriteHeader(http.StatusCreated)
				jsonResponse = jsonValidateResp{
					Body:      chirpResp.Body,
					ID:        chirpResp.ID,
					CreatedAt: chirpResp.CreatedAt,
					UpdatedAt: chirpResp.UpdatedAt,
					UserID:    chirpResp.UserID,
				}
			}

		}
		jsonOut, err := json.Marshal(jsonResponse)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			jsonResponse.Error = fmt.Sprintf("Something went wrong: %v", err)
		}
		w.Write(jsonOut)
	}
}

func (c *apiConfig) handlerGetChrips(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	userID, _ := uuid.Parse(req.URL.Query().Get("author_id"))
	sortOrder := req.URL.Query().Get("sort")
	var chirpList []database.Chirp
	var err error
	if userID == uuid.Nil {
		chirpList, err = c.dbQueries.GetAllChirps(req.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error retreving chirps: %v", err)))
			return
		}
	} else {
		chirpList, err = c.dbQueries.GetUsersChirps(req.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error retreving chirps: %v", err)))
			return
		}
	}

	if sortOrder == "desc" {
		sort.Slice(chirpList, func(a, b int) bool { return chirpList[a].CreatedAt.After(chirpList[b].CreatedAt) })
	} else {
		sort.Slice(chirpList, func(a, b int) bool { return chirpList[a].CreatedAt.Before(chirpList[b].CreatedAt) })
	}

	response := []jsonValidateResp{}
	for _, chirp := range chirpList {
		response = append(response, jsonValidateResp{
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			ID:        chirp.ID,
			UserID:    chirp.UserID,
		})
	}
	jsonOut, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error returning chirps: %v", err)))
	}
	w.Write(jsonOut)

}

func (c *apiConfig) handlerGetChirp(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var chirpOut jsonValidateResp
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		chirpOut.Error = fmt.Sprintf("error invalid UUID: %v", err)
		jsonOut, _ := json.Marshal(chirpOut)
		w.Write(jsonOut)
	} else {
		chirp, err := c.dbQueries.GetChirp(req.Context(), chirpID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			chirpOut.Error = fmt.Sprintf("error retreving chirp: %v", err)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Println(req.PathValue("chirpID"))
			chirpOut = jsonValidateResp{
				ID:        chirp.ID,
				Body:      chirp.Body,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				UserID:    chirp.UserID,
			}
		}
	}
	jsonOut, _ := json.Marshal(chirpOut)
	w.Write(jsonOut)
}

func (c *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	userID, err := auth.TokenAuth(req.Header, c.tokenSecret)
	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(fmt.Sprintf("Authorization error: %v", err)))
		return
	}
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Chirp ID invalid: %v", err)))
		return
	}
	chirp, err := c.dbQueries.GetChirp(req.Context(), chirpID)
	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Failed to reteve Chirp data: %v", err)))
		return
	}
	if userID == chirp.UserID {
		err = c.dbQueries.DeleteChirp(req.Context(), chirpID)
		if err != nil {
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Failed to delete Chirp: %v", err)))
			return
		} else {
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(fmt.Sprintf("Failed to delete Chirp: %v", err)))
			return
		}
	} else {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("User ID missmatch: %v", err)))
		return
	}
}
