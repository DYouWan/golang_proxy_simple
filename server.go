package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"proxy/config"
	"proxy/middleware"
	"strconv"
)

type Server struct {
	*config.Config
	//proxyMap 每一个路由对应一个反向代理，key存放的是 客户端路径模板
	proxyMap map[string]*Proxy
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		Config:   cfg,
		proxyMap: make(map[string]*Proxy, 0),
	}
}

func (c *Server) Start() error {
	router := mux.NewRouter()
	router.Use(middleware.PanicsHandling())
	router.Use(middleware.ElapsedTimeHandling())

	for _, route := range c.Routes {
		if err := c.ValidationAlgorithm(route.Algorithm); err != nil {
			return err
		}
		proxy, err := NewProxy(route)
		if err != nil {
			return err
		}
		if c.HealthCheck {
			proxy.HealthCheck(c.HealthCheckInterval)
		}
		c.proxyMap[route.UpstreamPathTemplate] = proxy

		router.Handle(route.UpstreamPathTemplate, proxy)//.Methods(route.UpstreamHTTPMethod...)
	}

	for key, proxy := range c.proxyMap {
		router.Handle(key, proxy)//.Methods(route.UpstreamHTTPMethod...)
	}
	if c.MaxAllowed > 0 {
		router.Use(middleware.MaxAllowedMiddleware(c.MaxAllowed))
	}
	svr := http.Server{
		Addr:    ":" + strconv.Itoa(c.Port),
		Handler: router,
	}

	// listen and serve
	if c.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else if c.Schema == "https" {
		err := svr.ListenAndServeTLS(c.CertCrt, c.CertKey)
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	}
	return nil
}


//func registerHost(w http.ResponseWriter, r *http.Request) {
//	_ = r.ParseForm()
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
