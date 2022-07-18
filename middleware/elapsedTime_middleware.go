package middleware

import (
	"net/http"
	"proxy/util"
	"proxy/util/logging"
	"time"
)

func ElapsedTimeHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		next.ServeHTTP(w, r)
		timeElapsed := time.Since(timeStart)
		logging.INFO.Println(util.GetIP(r), r.RequestURI, timeElapsed.String())
	})
}