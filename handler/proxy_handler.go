package handler

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxy/util"
	"strconv"
	"strings"
	"time"
)

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

//newSingleHostReverseProxy 获取下游主机ReverseProxy
func newSingleHostReverseProxy(targetUrl *url.URL,upstreamPath string,downstreamPath string)*httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Host = targetUrl.Host
		req.URL.Scheme = targetUrl.Scheme
		req.URL.Path = strings.Replace(req.URL.Path, upstreamPath, downstreamPath, 1)

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

