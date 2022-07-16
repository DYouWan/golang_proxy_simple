package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy/balancer"
	"proxy/config"
	"proxy/utils"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ReverseProxy = "Balancer-Reverse-Proxy"
)

type Proxy struct {
	bl              balancer.Balancer
	rw              sync.RWMutex
	alive           map[string]bool
	reverseProxyMap map[string]*httputil.ReverseProxy
}

// ServeHTTP implements a proxy to the http server
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host, err := p.bl.Balance(utils.GetHost(r.URL))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf("balance error: %s", err.Error())))
		return
	}
	p.bl.Inc(host)
	defer p.bl.Done(host)
	p.reverseProxyMap[host].ServeHTTP(w, r)
}

func NewProxy(routing config.Routing) (*Proxy,error) {
	var targetHosts []string
	alive := make(map[string]bool)
	reverseProxyMap := make(map[string]*httputil.ReverseProxy)
	for _, downstreamHost := range routing.DownstreamHostAndPorts {
		urlStr := fmt.Sprintf("%s://%s:%d%s", routing.DownstreamScheme, downstreamHost.Host, downstreamHost.Port, routing.DownstreamPathTemplate)
		parseUrl, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		host := utils.GetHost(parseUrl)
		reverseProxy := newSingleHostReverseProxy(routing.UpstreamPathTemplate, parseUrl)
		alive[host] = true
		reverseProxyMap[host] = reverseProxy

		targetHosts = append(targetHosts, host)
	}
	lb, err := balancer.Build(routing.Algorithm, targetHosts)
	if err != nil {
		return nil, err
	}

	proxy := &Proxy{bl: lb, alive: alive, reverseProxyMap: reverseProxyMap}
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

func newSingleHostReverseProxy(upstreamPathTemplate string,target *url.URL)*httputil.ReverseProxy {
	director := func(req *http.Request) {
		re, _ := regexp.Compile("^(.*){url}")
		targetPath := re.ReplaceAllString(target.Path, "$1")
		pathTemplate := re.ReplaceAllString(upstreamPathTemplate, "$1")
		reqPath := strings.Replace(req.URL.Path, pathTemplate, "", 1)

		req.URL.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Path = fmt.Sprintf("%s%s", targetPath, reqPath)

		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
		req.Header.Set(utils.XProxy, ReverseProxy)
		req.Header.Set(utils.XRealIP, utils.GetIP(req))
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

func (p *Proxy) HealthCheck(interval uint) {
	for host := range p.reverseProxyMap {
		go p.healthCheck(host, interval)
	}
}

func (p *Proxy) healthCheck(host string, interval uint) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		isBackendAlive := utils.IsBackendAlive(host)
		if !isBackendAlive && p.ReadAlive(host) {
			log.Printf("site unreachable, remove %s from load balancer.", host)

			p.SetAlive(host, false)
			p.bl.Remove(host)
		} else if isBackendAlive && !p.ReadAlive(host) {
			log.Printf("site reachable, add %s to load balancer.", host)

			p.SetAlive(host, true)
			p.bl.Add(host)
		}
	}
}

// ReadAlive reads the alive status of the site
func (p *Proxy) ReadAlive(url string) bool {
	p.rw.RLock()
	defer p.rw.RUnlock()
	return p.alive[url]
}

// SetAlive sets the alive status to the site
func (p *Proxy) SetAlive(url string, alive bool) {
	p.rw.Lock()
	defer p.rw.Unlock()
	p.alive[url] = alive
}
