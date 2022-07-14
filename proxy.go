package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"proxy/config"
	"proxy/middleware"
	"strconv"
	"time"
)

func ServerStart(configFile string) error {
	cfg, err := config.ReadConfig(configFile, true)
	if err != nil {
		return err
	}

	router := mux.NewRouter()
	//router.HandleFunc("/", health)
	//router.HandleFunc("/hello", hello)
	//router.HandleFunc("/helloWord", helloWord)
	for _, l := range cfg.Routing {
		httpProxy, err := NewHTTPProxy(l.ProxyPass, l.BalanceMode)
		if err != nil {
			log.Fatalf("create proxy error: %s", err)
		}
		// start health check
		if cfg.Server.HealthCheck {
			httpProxy.HealthCheck(cfg.Server.HealthCheckInterval)
		}
		router.Handle(l.Pattern, httpProxy)
	}
	if cfg.Server.MaxAllowed > 0 {
		router.Use(middleware.MaxAllowedMiddleware(cfg.Server.MaxAllowed))
	}
	svr := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: router,
	}

	// listen and serve
	if cfg.Server.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else if cfg.Server.Schema == "https" {
		err := svr.ListenAndServeTLS(cfg.Server.CertCrt, cfg.Server.CertKey)
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	}
	return nil
}



func health(w http.ResponseWriter,_ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "server is ok")
	time.Sleep(2e9) // 休眠2秒
	panic("health执行失败")
}
func hello(w http.ResponseWriter,_ *http.Request) {
	time.Sleep(2e9) // 休眠2秒
	w.Write([]byte("hello"))
}
func helloWord(w http.ResponseWriter,_ *http.Request) {
	time.Sleep(500) // 休眠2秒
	w.Write([]byte("hello word"))
}