package routing

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

//RouteMux 路由器存储多个路由
type RouteMux struct {
	mu              sync.RWMutex
	m               map[string]muxEntry
	middlewareChain []MiddlewareChainFunc
}

type MiddlewareChainFunc func(http.Handler) http.Handler

//muxEntry 存储路由对应的IRouteHandler
type muxEntry struct {
	h       http.Handler
	pattern string
}

func NewRouteMux() *RouteMux {
	return new(RouteMux)
}

// Handle 向路由器注册路由
func (mux *RouteMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("无效的请求路由")
	}
	if handler == nil {
		panic("无效的请求路由处理程序")
	}
	if _, exist := mux.m[pattern]; exist {
		panic("该请求路由已经被注册过 " + pattern)
	}

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}
	e := muxEntry{h: handler, pattern: pattern}
	mux.m[pattern] = e
}

func (mux *RouteMux) HandleFunc(pattern string, handler func(rw http.ResponseWriter,req *http.Request)) {
	if handler == nil {
		panic("处理程序不能为空")
	}
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// ServeTCP 根据请求数据的Header 找到对应的处理程序去执行
func (mux *RouteMux) ServeHTTP(rw http.ResponseWriter,req *http.Request) {
	h, _ := mux.Handler(req.RequestURI)
	if h == nil {
		rw.WriteHeader(http.StatusNotFound)
		str := fmt.Sprintf("%s%s", req.URL.RequestURI(), "当前页面不存在")
		rw.Write([]byte(str))
	} else {
		if len(mux.middlewareChain) > 0 {
			for i := range mux.middlewareChain {
				m := mux.middlewareChain[i](h)
				m.ServeHTTP(rw, req)
			}
		} else {
			h.ServeHTTP(rw, req)
		}
	}
}

// Handler 匹配处理程序
func (mux *RouteMux) Handler(header string) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	h, pattern = mux.match(header)
	if h == nil {
		v, ok := mux.m["default"]
		if ok {
			return v.h, v.pattern
		}
	}
	return
}

func (mux *RouteMux) match(path string) (h http.Handler, pattern string) {
	v, ok := mux.m[path]
	if ok {
		return v.h, v.pattern
	}else {
		for pathKey, entry := range mux.m {
			if len(pathKey) > len(path) {
				continue
			}
			if find := strings.Contains(path, pathKey); find {
				return entry.h, entry.pattern
			}
		}
	}
	return nil, ""
}

func (mux *RouteMux) Use(m MiddlewareChainFunc) {
	mux.middlewareChain = append(mux.middlewareChain, m)
}