package main

import "strings"

var badWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func replaceWords(text, mask string) string {
	words := strings.Split(text, " ")
	for i := range words {
		word := strings.ToLower(words[i])
		if _, ok := badWords[word]; ok {
			words[i] = mask
		}
	}
	return strings.Join(words, " ")
}
