package profanity

import "strings"

var bannedWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func Censor(text, mask string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			words[i] = mask
		}
	}
	return strings.Join(words, " ")
}
