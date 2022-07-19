package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"proxy/config"
	"proxy/middleware"
	"strconv"
)

func ServerStart(cfg *config.Config) error {
	muxRouter := mux.NewRouter()
	muxRouter.Use(middleware.ElapsedTimeHandling)
	muxRouter.Use(middleware.PanicsHandling)
	if cfg.MaxAllowed > 0 {
		muxRouter.Use(middleware.MaxAllowedMiddleware(cfg.MaxAllowed))
	}

	for _, r := range cfg.Routes {
		if err := cfg.ValidationAlgorithm(r.Algorithm); err != nil {
			return err
		}

		upstreamPath := r.UpstreamPathParse()
		downstreamPath := r.DownstreamPathParse()
		proxyRoute, err := NewProxyRoute(r.Algorithm, r.DownstreamScheme,upstreamPath, downstreamPath, r.DownstreamHostAndPorts)
		if err != nil {
			return err
		}

		if cfg.HealthCheck {
			proxyRoute.HealthCheck(cfg.HealthCheckInterval)
		}

		muxRouter.PathPrefix(upstreamPath).Handler(proxyRoute)
	}

	svr := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: muxRouter,
	}

	if cfg.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else {
		err := svr.ListenAndServeTLS(cfg.CertCrt, cfg.CertKey)
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	}
	return nil
}

//func (s *Server) RegisterHost(w http.ResponseWriter, r *http.Request)  {
//	_ = r.ParseForm()
//	host := r.Form["host"][0]
//
//
//	err := p.RegisterHost(r.Form["host"][0])
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		_, _ = fmt.Fprintf(w, err.Error())
//		return
//	}
//
//	_, _ = fmt.Fprintf(w, fmt.Sprintf("register host: %s success", r.Form["host"][0]))
//}



//func unregisterHost(w http.ResponseWriter, r *http.Request) {
//	_ = r.ParseForm()
//
//	err := p.UnregisterHost(r.Form["host"][0])
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		_, _ = fmt.Fprintf(w, err.Error())
//		return
//	}
//
//	_, _ = fmt.Fprintf(w, fmt.Sprintf("unregister host: %s success", r.Form["host"][0]))
//}
//
//func getKey(w http.ResponseWriter, r *http.Request) {
//	_ = r.ParseForm()
//
//	val, err := p.GetKey(r.Form["key"][0])
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		_, _ = fmt.Fprintf(w, err.Error())
//		return
//	}
//
//	_, _ = fmt.Fprintf(w, fmt.Sprintf("key: %s, val: %s", r.Form["key"][0], val))
//}
//
//func getKeyLeast(w http.ResponseWriter, r *http.Request) {
//	_ = r.ParseForm()
//
//	val, err := p.GetKeyLeast(r.Form["key"][0])
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		_, _ = fmt.Fprintf(w, err.Error())
//		return
//	}
//
//	_, _ = fmt.Fprintf(w, fmt.Sprintf("key: %s, val: %s", r.Form["key"][0], val))
//}