package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/auth"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/database"
)

type PolkaWebhook struct {
	Event string `json:"event,omitempty"`
	Data  struct {
		UserID uuid.UUID `json:"user_id,omitempty"`
	} `json:"data"`
}

func (c *apiConfig) handlerUserSetRed(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	var webhookData PolkaWebhook
	var message string
	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		message = fmt.Sprint(err)
	} else if apiKey != c.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		message = "Key missmatch"
	} else {
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&webhookData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else if webhookData.Event == "user.upgraded" {
			var params database.AddRedtoUserParams
			params.ID = webhookData.Data.UserID
			params.IsChirpyRed = true
			_, err := c.dbQueries.AddRedtoUser(req.Context(), params)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
	w.Write([]byte(message))
}
