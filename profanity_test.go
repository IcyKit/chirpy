package main

import "testing"

func TestReplaceWords(t *testing.T) {
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
		if got := replaceWords(tt.in, "****"); got != tt.want {
			t.Errorf("replaceWords(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
