package middleware

import (
	"github.com/gorilla/mux"
	"net/http"
	"proxy/basis/logging"
)

func LogPanics() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if error := recover(); error != nil {
					logging.ERROR.Printf("[%v] caught panic: %v", r.RemoteAddr, error)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
