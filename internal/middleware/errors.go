package middleware

import (
	"log"
	"net/http"

	"github.com/francisco/gridironmind/pkg/response"
)

// RecoverPanic recovers from panics and returns a 500 error
func RecoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				response.InternalError(w, "An unexpected error occurred")
			}
		}()
		next(w, r)
	}
}

// LogRequest logs all incoming requests
func LogRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}