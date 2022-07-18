package config

type Routing struct {
	//UpstreamHTTPMethod 表示客户端请求到代理时，所允许HTTP请求的方法
	UpstreamHTTPMethod []string `json:"UpstreamHttpMethod"`
	//UpstreamPathTemplate 客户端请求代理时的Url路径模板
	UpstreamPathTemplate string `json:"UpstreamPathTemplate"`
	//DownstreamScheme 请求协议，目前支持http和https
	DownstreamScheme string `json:"DownstreamScheme"`
	//Algorithm 使用的负载均衡算法
	Algorithm string `json:"Algorithm"`
	//UseServiceDiscovery 是否启用服务发现
	UseServiceDiscovery bool `json:"UseServiceDiscovery"`
	//DownstreamPathTemplate 代理向目标转发时的Url路径模板
	DownstreamPathTemplate string `json:"DownstreamPathTemplate"`
	//DownstreamHostAndPorts 代理向下游转发地址集合
	DownstreamHostAndPorts []DownstreamHost `json:"DownstreamHostAndPorts"`
}

type DownstreamHost struct {
	Host string `json:"Host"`
	Port int    `json:"Port"`
}