package main

import (
	"sync/atomic"
	"time"

	"github.com/icykit/chirpy/internal/database"
)

const (
	staticDir       = "./public"
	maxChirpLength  = 140
	accessTokenTTL  = time.Hour
	refreshTokenTTL = 60 * 24 * time.Hour
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}
