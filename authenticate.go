package main

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/icykit/chirpy/internal/auth"
)

// Достает JWT из заголовка Authorization и возвращает id пользователя
func (cfg *apiConfig) authenticate(r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.UUID{}, err
	}
	return auth.ValidateJWT(token, cfg.jwtSecret)
}
