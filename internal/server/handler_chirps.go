package server

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/icykit/chirpy/internal/database"
	"github.com/icykit/chirpy/internal/profanity"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type chirpRequest struct {
	Body string `json:"body"`
}

func toChirp(c database.Chirp) Chirp {
	return Chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}
}

func (s *Server) handlerChirpsList(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortDir := r.URL.Query().Get("sort")

	var dbChirps []database.Chirp
	var err error

	if authorID == "" {
		dbChirps, err = s.db.GetChirps(r.Context())
	} else {
		authorUUID, parseErr := uuid.Parse(authorID)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id", nil)
			return
		}
		dbChirps, err = s.db.GetChirpsByUser(r.Context(), authorUUID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	chirps := make([]Chirp, 0, len(dbChirps))
	for _, c := range dbChirps {
		chirps = append(chirps, toChirp(c))
	}

	if sortDir == "desc" {
		slices.Reverse(chirps)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (s *Server) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", nil)
		return
	}

	chirp, err := s.db.GetChirp(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp", err)
		return
	}
	respondWithJSON(w, http.StatusOK, toChirp(chirp))
}

func (s *Server) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	userID, err := s.authenticate(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	var req chirpRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request body", nil)
		return
	}

	if utf8.RuneCountInString(req.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp, err := s.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   profanity.Censor(req.Body, "****"),
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}
	respondWithJSON(w, http.StatusCreated, toChirp(chirp))
}

func (s *Server) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	userID, err := s.authenticate(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", nil)
		return
	}

	chirp, err := s.db.GetChirp(r.Context(), id)
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

	if err := s.db.DeleteChirp(r.Context(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
