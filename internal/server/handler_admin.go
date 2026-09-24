package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", s.fileserverHits.Load())
}

func (s *Server) handlerReset(w http.ResponseWriter, r *http.Request) {
	if s.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	if err := s.db.DeleteUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't reset database", err)
		return
	}

	s.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
