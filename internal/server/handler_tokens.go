package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/icykit/chirpy/internal/auth"
)

func (s *Server) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	user, err := s.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user for refresh token", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, s.jwtSecret, accessTokenTTL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: token})
}

func (s *Server) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	if err := s.db.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
