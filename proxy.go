package main

import (
	"github.com/gorilla/mux"
	"proxy/config"
	"proxy/middleware"
)

func ServerStart(cfg *config.Config) error {
	router := mux.NewRouter()
	router.Use(middleware.PanicsHandling())
	router.Use(middleware.ElapsedTimeHandling())

	for _, route := range cfg.Routes {
		if err := cfg.ValidationAlgorithm(route.Algorithm);err!= nil {
			return err
		}
		//for i, i2 := range collection {
		//
		//}
		//
		//lb, err := balancer.Build(route.Algorithm, route.)
	}
	return nil
}


//
//func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
//	//http://127.0.0.1:2002/dir?name=123
//	//RayQuery: name=123
//	//Scheme: http
//	//Host: 127.0.0.1:2002
//	targetQuery := target.RawQuery
//	director := func(req *http.Request) {
//		//url_rewrite
//		//127.0.0.1:2002/dir/abc ==> 127.0.0.1:2003/base/abc ??
//		//127.0.0.1:2002/dir/abc ==> 127.0.0.1:2002/abc
//		//127.0.0.1:2002/abc ==> 127.0.0.1:2003/base/abc
//		re, _ := regexp.Compile("^/dir(.*)");
//		req.URL.Path = re.ReplaceAllString(req.URL.Path, "$1")
//
//		req.URL.Scheme = target.Scheme
//		req.URL.Host = target.Host
//
//		//target.Path : /base
//		//req.URL.Path : /dir
//		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
//		if targetQuery == "" || req.URL.RawQuery == "" {
//			req.URL.RawQuery = targetQuery + req.URL.RawQuery
//		} else {
//			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
//		}
//		if _, ok := req.Header["User-Agent"]; !ok {
//			req.Header.Set("User-Agent", "")
//		}
//	}
//
//	return nil
//}