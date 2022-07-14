package middleware

import (
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"proxy/basis/logging"
	"time"
)

func TimeMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timeStart := time.Now()
			next.ServeHTTP(w,r)
			timeElapsed := time.Since(timeStart)
			logging.INFO.Println(getRemoteClientIp(r), r.RequestURI,timeElapsed.String())
		})
	}
}

// GetRemoteClientIp 获取远程客户端IP
func getRemoteClientIp(r *http.Request) string {
	remoteIp := r.RemoteAddr
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		remoteIp = ip
	} else if ip = r.Header.Get("X-Forwarded-For"); ip != "" {
		remoteIp = ip
	} else {
		remoteIp, _, _ = net.SplitHostPort(remoteIp)
	}
	//本地ip
	if remoteIp == "::1" {
		remoteIp = "127.0.0.1"
	}
	return remoteIp
}