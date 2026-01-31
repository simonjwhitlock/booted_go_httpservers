package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/database"
)

type PolkaWebhook struct {
	Event string `json:"event,omitempty"`
	Data  struct {
		UserID uuid.UUID `json:"user_id,omitempty"`
	} `json:"data"`
}

func (c *apiConfig) handlerUserSetRed(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var webhookData PolkaWebhook
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
	w.Write([]byte(""))
}
