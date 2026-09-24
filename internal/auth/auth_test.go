package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJwt(t *testing.T) {
	userID := uuid.New()
	secret := "secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT: %v", err)
	}

	got, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if got != userID {
		t.Errorf("got %v, want %v", got, userID)
	}

	if _, err := ValidateJWT(token, "wrong-secret"); err == nil {
		t.Error("expected error for wrong secret")
	}

	expired, _ := MakeJWT(userID, secret, -time.Hour)
	if _, err := ValidateJWT(expired, secret); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{"valid", "Bearer abc123", "abc123", false},
		{"missing header", "", "", true},
		{"no space", "abc123", "", true},
		{"wrong scheme", "ApiKey abc123", "", true},
		{"empty token", "Bearer ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.header != "" {
				headers.Set("Authorization", tt.header)
			}

			got, err := GetBearerToken(headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "ApiKey secret-key")

	got, err := GetAPIKey(headers)
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if got != "secret-key" {
		t.Errorf("got %q, want %q", got, "secret-key")
	}

	headers.Set("Authorization", "Bearer secret-key")
	if _, err := GetAPIKey(headers); err == nil {
		t.Error("expected error for Bearer scheme")
	}
}
