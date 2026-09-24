package server

import "net/http"

func (s *Server) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
