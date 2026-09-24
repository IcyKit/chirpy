package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "chirpy-access"

var (
	ErrNoAuthHeader        = errors.New("authorization header is missing")
	ErrMalformedAuthHeader = errors.New("malformed authorization header")
)

// Хеширует пароль
//
// В качестве аргумента принимает сам пароль (строка)
func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Проверка хеша пароля на действительность
func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

// Создает JWT токен
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	method := jwt.SigningMethodHS256
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(method, claims)
	return token.SignedString([]byte(tokenSecret))
}

// Валидирует JWT токен
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	validatedJwt, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
	)
	if err != nil {
		return uuid.UUID{}, err
	}

	if !validatedJwt.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid subject: %w", err)
	}

	return userID, nil
}

// Получает авторизационный токен пользователя
func GetBearerToken(headers http.Header) (string, error) {
	return getAuthorization(headers, "Bearer ")
}

func GetAPIKey(headers http.Header) (string, error) {
	return getAuthorization(headers, "ApiKey ")
}

func getAuthorization(headers http.Header, prefix string) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", ErrNoAuthHeader
	}

	value, ok := strings.CutPrefix(header, prefix)
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return "", ErrMalformedAuthHeader
	}
	return value, nil
}

// Создает refresh токен
func MakeRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
