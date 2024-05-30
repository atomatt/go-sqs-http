package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func newHandler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, map[string]string{
			"hello": "world",
		})
	})

	return mux
}

func jsonOK(w http.ResponseWriter, body any) {
	jsonResponse(w, 200, body)
}

func jsonResponse(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error(err.Error())
	}
}
