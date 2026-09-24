package main

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/icykit/chirpy/internal/database"
)

type chirpReq struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) handlerChirpsList(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortDir := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error

	if authorID == "" {
		chirps, err = cfg.db.GetChirps(r.Context())
	} else {
		authorUUID, parseErr := uuid.Parse(authorID)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id", nil)
			return
		}
		chirps, err = cfg.db.GetChirpsByUser(r.Context(), authorUUID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	// Без этого пустой список сериализуется в null, а не в []
	if chirps == nil {
		chirps = []database.Chirp{}
	}

	// Запросы уже сортируют по created_at по возрастанию
	if sortDir == "desc" {
		slices.Reverse(chirps)
	}

	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", nil)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp", err)
		return
	}
	respondWithJson(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.authenticate(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	var c chirpReq
	if err := decodeJSON(w, r, &c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request body", nil)
		return
	}

	if utf8.RuneCountInString(c.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedBody := replaceWords(c.Body, "****")
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanedBody, UserID: userID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}
	respondWithJson(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	userID, err := cfg.authenticate(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	id, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", nil)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp", err)
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
