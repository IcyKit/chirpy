package main

import "net/http"

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	respondWithJson(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "OK"})
}
