package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/icykit/chirpy/internal/auth"
	"github.com/icykit/chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type userReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func toUser(u database.User) User {
	return User{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Email:       u.Email,
		IsChirpyRed: u.IsChirpyRed,
	}
}

// Проверяет, что ошибка — нарушение UNIQUE (например, email уже занят)
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	var u userReq
	if err := decodeJSON(w, r, &u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request body", nil)
		return
	}

	hash, err := auth.HashPassword(u.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: u.Email, HashedPassword: hash})
	if isUniqueViolation(err) {
		respondWithError(w, http.StatusConflict, "Email is already in use", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJson(w, http.StatusCreated, toUser(user))
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	var u userReq
	if err := decodeJSON(w, r, &u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request body", nil)
		return
	}

	hashedPass, err := auth.HashPassword(u.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          u.Email,
		HashedPassword: hashedPass,
		ID:             userID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "User not found", nil)
		return
	}
	if isUniqueViolation(err) {
		respondWithError(w, http.StatusConflict, "Email is already in use", nil)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	resp := toUser(user)
	resp.Token = token
	respondWithJson(w, http.StatusOK, resp)
}
