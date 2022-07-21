package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy/balancer"
	"proxy/util/logging"
	"strings"
	"sync"
)

var (
	ReverseProxy = "Balancer-Reverse-Proxy"
)

//RoutePrefixHandler 前缀路由处理程序
type RoutePrefixHandler struct {
	mux sync.RWMutex
	//bl 通过请求时的url，获取具体的负载均衡器
	bl balancer.Balancer
	//UpstreamPath 上游请求路径
	UpstreamPath string
	//DownstreamPath 下游请求路径
	DownstreamPath string
	//alive 主机存活检测
	alive map[string]bool
	//reverseProxyMap 根据负载均衡器返回的host，获取对应的反向代理
	reverseProxyMap map[string]*httputil.ReverseProxy
	//builtinHandler 内置处理程序
	builtinHandler map[string]func(w http.ResponseWriter, r *http.Request)
}

//NewRoutePrefixHandler 接收下游的主机信息，返回下游主机代理
func NewRoutePrefixHandler(algorithm string,upstreamPath string,downstreamPath string, downstreamHosts []string) (*RoutePrefixHandler,error) {
	var targetHosts []string
	alive := make(map[string]bool)
	reverseProxyMap := make(map[string]*httputil.ReverseProxy)

	for _, dh := range downstreamHosts {
		dest, err := url.Parse(dh)
		if err != nil || dest.Scheme == "" || dest.Host == "" {
			return nil, err
		}
		host := cleanHost(dest.Host)
		alive[host] = true
		targetHosts = append(targetHosts, host)
		reverseProxyMap[host] = newSingleHostReverseProxy(dest, upstreamPath, downstreamPath)

		logging.Infof("主机 %s 初始化成功", dh)
	}
	bl, err := balancer.Build(algorithm, targetHosts)
	if err != nil {
		return nil, err
	}

	prefixHandler := &RoutePrefixHandler{
		bl:              bl,
		alive:           alive,
		UpstreamPath:    upstreamPath,
		DownstreamPath:  downstreamPath,
		reverseProxyMap: reverseProxyMap,
	}
	prefixHandler.builtinHandler = map[string]func(w http.ResponseWriter, r *http.Request){
		upstreamPath + "/register":   prefixHandler.registerHost,
		upstreamPath + "/unregister": prefixHandler.unregisterHost,
	}
	return prefixHandler, nil
}

//ServeHTTP 实现到http服务器的代理
func (rh *RoutePrefixHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//当前缀匹配进来之后，先判断是否请求内置的接口
	handler := rh.builtinHandler[r.URL.Path]
	if handler != nil {
		handler(w, r)
		return
	}
	//如果不是请求内置接口，则进行转发
	key := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
	host, err := rh.bl.Balance(key)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		errStr := fmt.Sprintf("负载均衡器: %s", err.Error())
		logging.Error(errStr)
		_, _ = w.Write([]byte(errStr))
		return
	}
	rh.bl.Inc(host)
	defer rh.bl.Done(host)
	rh.reverseProxyMap[host].ServeHTTP(w, r)
}

func cleanHost(in string) string {
	if i := strings.IndexAny(in, " /"); i != -1 {
		return in[:i]
	}
	return in
}