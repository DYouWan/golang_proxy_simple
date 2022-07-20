package middleware

import (
	"net/http"
	"proxy/util/logging"
)

func PanicsHandling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logging.Errorf("[%v]请求%s?%s 异常: %v", r.RemoteAddr, r.URL.Path, r.URL.RawQuery, err)
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(err.(error).Error()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}