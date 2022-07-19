package route

import (
	"log"
	"net/http"
	"proxy/config"
	"proxy/middleware"
	"strconv"
)

type ServerRoute struct {
	*config.Config

	//proxyMap 路由Or中间件处理器
	routeMux *MuxRoute

	//proxyMap 每一个路由对应一个反向代理，key存放的是 客户端路径模板
	proxyMap map[string]*ProxyRoute
}

func NewServerRoute(cfg *config.Config) *ServerRoute {
	return &ServerRoute{
		Config:   cfg,
		routeMux: NewMuxRoute(),
		proxyMap: make(map[string]*ProxyRoute, 0),
	}
}

func (sr *ServerRoute) Start() error {
	sr.routeMux.Use(middleware.PanicsHandling)
	sr.routeMux.Use(middleware.ElapsedTimeHandling)
	for _, r := range sr.Routes {
		if err := sr.ValidationAlgorithm(r.Algorithm); err != nil {
			return err
		}

		downstreamPath := r.DownstreamPathParse()
		proxyRoute, err := NewProxyRoute(r.Algorithm, r.DownstreamScheme, downstreamPath, r.DownstreamHostAndPorts)
		if err != nil {
			return err
		}

		if sr.HealthCheck {
			proxyRoute.HealthCheck(sr.HealthCheckInterval)
		}

		upstreamPath := r.UpstreamPathParse()
		sr.proxyMap[upstreamPath] = proxyRoute
		sr.routeMux.Handle(upstreamPath, proxyRoute)
	}

	if sr.MaxAllowed > 0 {
		sr.routeMux.Use(middleware.MaxAllowedMiddleware(sr.MaxAllowed))
	}
	svr := http.Server{
		Addr:    ":" + strconv.Itoa(sr.Port),
		Handler: sr.routeMux,
	}

	// listen and serve
	if sr.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else if sr.Schema == "https" {
		err := svr.ListenAndServeTLS(sr.CertCrt, sr.CertKey)
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