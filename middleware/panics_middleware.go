package middleware

import (
	"net/http"
	"proxy/util/logging"
)

func PanicsHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logging.ERROR.Printf("[%v] proxy panic: %v", r.RemoteAddr, err)
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(err.(error).Error()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}