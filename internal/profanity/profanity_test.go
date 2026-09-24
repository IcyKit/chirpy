package profanity

import "testing"

func TestCensor(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello world", "hello world"},
		{"what a kerfuffle", "what a ****"},
		{"Sharbert FORNAX ok", "**** **** ok"},
		{"kerfuffle!", "kerfuffle!"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := Censor(tt.in, "****"); got != tt.want {
			t.Errorf("Censor(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
