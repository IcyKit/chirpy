package server

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/icykit/chirpy/internal/config"
	"github.com/icykit/chirpy/internal/database"
)

const (
	maxChirpLength  = 140
	accessTokenTTL  = time.Hour
	refreshTokenTTL = 60 * 24 * time.Hour
)

type Server struct {
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
	staticDir      string
	fileserverHits atomic.Int32
}

func New(db *database.Queries, cfg config.Config) *Server {
	return &Server{
		db:        db,
		platform:  cfg.Platform,
		jwtSecret: cfg.JWTSecret,
		polkaKey:  cfg.PolkaKey,
		staticDir: cfg.StaticDir,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(s.staticDir)))
	mux.Handle("/app/", s.middlewareMetricsInc(fileServer))

	mux.HandleFunc("GET /api/healthz", handlerHealthz)

	mux.HandleFunc("GET /admin/metrics", s.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", s.handlerReset)

	mux.HandleFunc("POST /api/users", s.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", s.handlerUsersUpdate)
	mux.HandleFunc("POST /api/login", s.handlerLogin)
	mux.HandleFunc("POST /api/refresh", s.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", s.handlerRevoke)

	mux.HandleFunc("GET /api/chirps", s.handlerChirpsList)
	mux.HandleFunc("GET /api/chirps/{chirpID}", s.handlerChirpsGet)
	mux.HandleFunc("POST /api/chirps", s.handlerChirpsCreate)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", s.handlerChirpsDelete)

	mux.HandleFunc("POST /api/polka/webhooks", s.handlerPolkaWebhook)

	return mux
}
