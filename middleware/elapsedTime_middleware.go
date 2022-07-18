package middleware

import (
	"github.com/gorilla/mux"
	"net/http"
	"proxy/util"
	logging2 "proxy/util/logging"
	"time"
)

func ElapsedTimeHandling() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timeStart := time.Now()
			next.ServeHTTP(w,r)
			timeElapsed := time.Since(timeStart)
			logging2.INFO.Println(util.GetIP(r), r.RequestURI,timeElapsed.String())
		})
	}
}