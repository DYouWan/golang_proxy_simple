package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"proxy/util/logging"
)

//var (
//	NoHostError                = errors.New("no host")
//	ErrHostNotFound            = errors.New("host not found")
//	InvalidTargetHost          = errors.New("invalid target host")
//	ErrHostAlreadyExists       = errors.New("host already exists")
//	AlgorithmNotSupportedError = errors.New("algorithm not supported")
//)

//registerHost 添加主机
func (rh *RoutePrefixHandler) registerHost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	var resultStr string
	urlStr := r.Form["url"][0]
	dest, err := url.Parse(urlStr)
	if err != nil || dest.Scheme == "" || dest.Host == "" {
		resultStr = fmt.Sprintf("无效的主机: %s", urlStr)
	}else {
		host := cleanHost(dest.Host)
		if rh.reverseProxyMap[host] != nil {
			resultStr = fmt.Sprintf("主机: %s 已存在", urlStr)
		}else {
			rh.bl.Add(host)
			rh.alive[host] = true
			rh.reverseProxyMap[host] = newSingleHostReverseProxy(dest, rh.UpstreamPath, rh.DownstreamPath)
			logging.Infof("主机 %s 初始化成功", urlStr)
			resultStr = fmt.Sprintf("主机 %s 初始化成功", urlStr)
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(resultStr))
}

func (rh *RoutePrefixHandler) unregisterHost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	host := r.Form["host"][0]
	if host != "" {
		rh.bl.Remove(host)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("主机: %s 删除成功", host)))
}
