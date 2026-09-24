package main

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/icykit/chirpy/internal/auth"
)

type polkaReq struct {
	Event string `json:"event"`
	Data  struct {
		UserId uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(cfg.polkaKey)) != 1 {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", nil)
		return
	}

	var req polkaReq
	if err := decodeJSON(w, r, &req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request body", nil)
		return
	}

	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpgradeUser(r.Context(), req.Data.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "User not found", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upgrade user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
