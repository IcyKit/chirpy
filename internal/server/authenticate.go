package server

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/icykit/chirpy/internal/auth"
)

func (s *Server) authenticate(r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.UUID{}, err
	}
	return auth.ValidateJWT(token, s.jwtSecret)
}
