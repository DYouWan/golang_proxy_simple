package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy/balancer"
	"proxy/config"
	"proxy/middleware"
	"strconv"
)

func ServerStart(configFile string) error {
	cfg, err := config.ReadConfig(configFile, true)
	if err != nil {
		return err
	}

	router := mux.NewRouter()
	router.Use(middleware.PanicsHandling())
	router.Use(middleware.ElapsedTimeHandling())

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

func NewHTTPProxy2(algorithm string,targetHosts []string) (*HTTPProxy, error) {
	balancer, err := balancer.Build(algorithm, targetHosts)
	if err != nil {
		return nil, err
	}
	fmt.Println(balancer)
	return nil, nil
}

func newSingleHostReverseProxy(targetHost *url.URL) *httputil.ReverseProxy {
	return nil
}