package server

import (
	"encoding/json"
	"net/http"
)

type apiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type appHandler func(w http.ResponseWriter, r *http.Request) *apiError

// Method allows appHandler to satisfy the http.Handler interface.
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Status)
		json.NewEncoder(w).Encode(err)
	}
}
