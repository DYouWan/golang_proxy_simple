package config

import (
	"fmt"
	"regexp"
	"strings"
)

type Routing struct {
	//UpstreamHTTPMethod 表示客户端请求到代理时，所允许HTTP请求的方法
	UpstreamHTTPMethod []string `json:"UpstreamHttpMethod"`
	//UpstreamPathTemplate 客户端请求代理时的Url路径模板
	UpstreamPathTemplate string `json:"UpstreamPathTemplate"`
	//Algorithm 使用的负载均衡算法
	Algorithm string `json:"Algorithm"`
	//UseServiceDiscovery 是否启用服务发现
	UseServiceDiscovery bool `json:"UseServiceDiscovery"`
	//DownstreamPathTemplate 代理向目标转发时的Url路径模板
	DownstreamPathTemplate string `json:"DownstreamPathTemplate"`
	//DownstreamHostAndPorts 代理向下游转发地址集合
	DownstreamHosts []string `json:"DownstreamHosts"`
}

//ValidationAlgorithm 验证算法是否支持
func (r *Routing) ValidationAlgorithm() error {
	var exists bool
	algorithms := strings.Split(Algorithms, "|")
	for _, algorithm := range algorithms {
		if algorithm == r.Algorithm {
			exists = true
			break
		}
	}
	if exists == false {
		return fmt.Errorf("该 \"%s\" 算法不支持", r.Algorithm)
	}
	return nil
}

//UpstreamPathParse 上游路径解析
func (r *Routing) UpstreamPathParse() string {
	//验证是否以/开头
	matched, _ := regexp.MatchString("^/.*", r.UpstreamPathTemplate)
	if !matched {
		panic(fmt.Errorf("客户端上游请求路径模板 \"%s\" 不正确 ", r.UpstreamPathTemplate))
	}
	//验证是否存在占位符
	matched, _ = regexp.MatchString(".*{url}$", r.UpstreamPathTemplate)
	if !matched {
		return r.UpstreamPathTemplate
	}
	//获取占位符之前的路径
	re, _ := regexp.Compile("^(.*)/{url}$")
	prefixPath := re.ReplaceAllString(r.UpstreamPathTemplate, "$1")
	return prefixPath
}

//DownstreamPathParse 下游路径解析
func (r *Routing) DownstreamPathParse() string {
	//验证是否以/开头
	matched, _ := regexp.MatchString("^/.*", r.DownstreamPathTemplate)
	if !matched {
		panic(fmt.Errorf("客户端下游请求路径模板 \"%s\" 不正确 ", r.DownstreamPathTemplate))
	}
	//验证是否存在占位符
	matched, _ = regexp.MatchString(".*{url}$", r.DownstreamPathTemplate)
	if !matched {
		return r.DownstreamPathTemplate
	}
	//获取占位符之前的路径
	re, _ := regexp.Compile("^(.*)/{url}$")
	prefixPath := re.ReplaceAllString(r.DownstreamPathTemplate, "$1")
	return prefixPath
}