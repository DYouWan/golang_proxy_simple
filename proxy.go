package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"proxy/balancer"
	"proxy/config"
	"proxy/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ReverseProxy = "Balancer-Reverse-Proxy"
)

//ProxyRoute 反向代理
type ProxyRoute struct {
	mux sync.RWMutex
	//bl 通过请求时的url，获取具体的负载均衡器
	bl balancer.Balancer
	//alive 主机存活检测
	alive map[string]bool
	//reverseProxyMap 根据负载均衡器返回的host，获取对应的反向代理
	reverseProxyMap map[string]*httputil.ReverseProxy
}

//ServeHTTP 实现到http服务器的代理
func (p *ProxyRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery)
	host, err := p.bl.Balance(key)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf("balance error: %s", err.Error())))
		return
	}
	p.bl.Inc(host)
	defer p.bl.Done(host)
	p.reverseProxyMap[host].ServeHTTP(w, r)
}

//HealthCheck 主机健康检查
func (p *ProxyRoute) HealthCheck(interval uint) {
	for host := range p.reverseProxyMap {
		go p.healthCheck(host, interval)
	}
}

func (p *ProxyRoute) healthCheck(host string, interval uint) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		isBackendAlive := util.IsBackendAlive(host)
		if !isBackendAlive && p.ReadAlive(host) {
			log.Printf("该主机 %s 不可用，已经从负载均衡器中移除", host)

			p.SetAlive(host, false)
			p.bl.Remove(host)
		} else if isBackendAlive && !p.ReadAlive(host) {
			log.Printf("该主机 %s 正常，已添加到负载均衡器", host)

			p.SetAlive(host, true)
			p.bl.Add(host)
		}
	}
}

// ReadAlive 获取主机存活状态
func (p *ProxyRoute) ReadAlive(url string) bool {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.alive[url]
}

// SetAlive 设置主机存活状态
func (p *ProxyRoute) SetAlive(url string, alive bool) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.alive[url] = alive
}

//NewProxyRoute 接收下游的主机信息，返回下游主机代理
func NewProxyRoute(algorithm string,scheme string,upstreamPath string,downstreamPath string, downstreamHosts []config.DownstreamHost) (*ProxyRoute,error) {
	var targetHosts []string
	alive := make(map[string]bool)
	reverseProxyMap := make(map[string]*httputil.ReverseProxy)

	for _, dsh := range downstreamHosts {
		host, err := dsh.GetDownstreamHost(scheme)
		if err != nil {
			return nil, err
		}
		alive[host] = true
		targetHosts = append(targetHosts, host)
		reverseProxyMap[host] = newSingleHostReverseProxy(scheme, host, upstreamPath, downstreamPath)
	}
	lb, err := balancer.Build(algorithm, targetHosts)
	if err != nil {
		return nil, err
	}

	proxy := &ProxyRoute{
		bl:              lb,
		alive:           alive,
		reverseProxyMap: reverseProxyMap,
	}
	return proxy, nil
}

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, //连接超时
		KeepAlive: 30 * time.Second, //长连接超时时间
	}).DialContext,
	MaxIdleConns:          100,              //最大空闲连接
	IdleConnTimeout:       90 * time.Second, //空闲超时时间
	TLSHandshakeTimeout:   10 * time.Second, //tls握手超时时间
	ExpectContinueTimeout: 1 * time.Second,  //100-continue 超时时间
}

func newSingleHostReverseProxy(scheme string,host string,upstreamPath string,downstreamPath string)*httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Host = host
		req.URL.Scheme = scheme

		targetPath := strings.Replace(req.URL.Path, upstreamPath, downstreamPath, 1)
		req.URL.Path = targetPath

		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
		req.Header.Set(util.XProxy, ReverseProxy)
		req.Header.Set(util.XRealIP, util.GetIP(req))
	}

	//更改内容
	modifyFunc := func(resp *http.Response) error {
		if resp.StatusCode != 200 {
			//获取内容
			oldPayload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			//追加内容
			newPayload := []byte("StatusCode error:" + string(oldPayload))
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(newPayload))
			resp.ContentLength = int64(len(newPayload))
			resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(newPayload)), 10))
		}
		return nil
	}

	//错误回调 ：关闭real_server时测试，错误回调
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "ErrorHandler error:"+err.Error(), 500)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		Transport:      transport,
		ModifyResponse: modifyFunc,
		ErrorHandler:   errorHandler,
	}
}