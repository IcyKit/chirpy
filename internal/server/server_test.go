package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/icykit/chirpy/internal/auth"
	"github.com/icykit/chirpy/internal/config"
)

const testSecret = "test-secret"

func newTestServer(t *testing.T) *Server {
	t.Helper()

	root := t.TempDir()
	staticDir := filepath.Join(root, "public")
	if err := os.Mkdir(staticDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=leak"), 0o644); err != nil {
		t.Fatal(err)
	}

	return New(nil, config.Config{
		Platform:  "prod",
		JWTSecret: testSecret,
		PolkaKey:  "polka-key",
		StaticDir: staticDir,
	})
}

func TestEndpointsWithoutDatabase(t *testing.T) {
	validToken, err := auth.MakeJWT(uuid.New(), testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		authHeader string
		wantStatus int
		wantError  string
	}{
		{
			name:       "healthz",
			method:     http.MethodGet,
			path:       "/api/healthz",
			wantStatus: http.StatusOK,
		},
		{
			name:       "reset is forbidden outside dev",
			method:     http.MethodPost,
			path:       "/admin/reset",
			wantStatus: http.StatusForbidden,
			wantError:  "Forbidden",
		},
		{
			name:       "create chirp without token",
			method:     http.MethodPost,
			path:       "/api/chirps",
			body:       `{"body":"hi"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "create chirp with malformed header",
			method:     http.MethodPost,
			path:       "/api/chirps",
			authHeader: "abc",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "create chirp with token signed by another secret",
			method:     http.MethodPost,
			path:       "/api/chirps",
			authHeader: "Bearer " + mustMakeJWT(t, "other-secret"),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "create chirp with invalid JSON",
			method:     http.MethodPost,
			path:       "/api/chirps",
			body:       `{oops`,
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusBadRequest,
			wantError:  "Couldn't decode request body",
		},
		{
			name:       "create chirp that is too long",
			method:     http.MethodPost,
			path:       "/api/chirps",
			body:       `{"body":"` + strings.Repeat("a", maxChirpLength+1) + `"}`,
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusBadRequest,
			wantError:  "Chirp is too long",
		},
		{
			name:       "get chirp with invalid ID",
			method:     http.MethodGet,
			path:       "/api/chirps/not-a-uuid",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid chirp ID",
		},
		{
			name:       "list chirps with invalid author_id",
			method:     http.MethodGet,
			path:       "/api/chirps?author_id=nope",
			wantStatus: http.StatusBadRequest,
			wantError:  "Invalid author_id",
		},
		{
			name:       "login with invalid JSON",
			method:     http.MethodPost,
			path:       "/api/login",
			body:       `{oops`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "webhook with Bearer instead of ApiKey",
			method:     http.MethodPost,
			path:       "/api/polka/webhooks",
			authHeader: "Bearer polka-key",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "webhook with wrong key",
			method:     http.MethodPost,
			path:       "/api/polka/webhooks",
			authHeader: "ApiKey wrong",
			wantStatus: http.StatusUnauthorized,
			wantError:  "Invalid API key",
		},
		{
			name:       "webhook ignores unknown events",
			method:     http.MethodPost,
			path:       "/api/polka/webhooks",
			body:       `{"event":"user.downgraded"}`,
			authHeader: "ApiKey polka-key",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "static index",
			method:     http.MethodGet,
			path:       "/app/",
			wantStatus: http.StatusOK,
		},
		{
			name:       "files outside the static dir are not served",
			method:     http.MethodGet,
			path:       "/app/.env",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestServer(t).Handler()

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body)
			}

			if tt.wantError != "" {
				var resp errorResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				if resp.Error != tt.wantError {
					t.Errorf("error = %q, want %q", resp.Error, tt.wantError)
				}
			}
		})
	}
}

func TestMetricsCountsFileserverHits(t *testing.T) {
	handler := newTestServer(t).Handler()

	for range 3 {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/app/", nil))
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/metrics", nil))

	if !strings.Contains(rec.Body.String(), "visited 3 times") {
		t.Errorf("metrics body = %q, want it to mention 3 visits", rec.Body)
	}
}

func mustMakeJWT(t *testing.T, secret string) string {
	t.Helper()
	token, err := auth.MakeJWT(uuid.New(), secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
