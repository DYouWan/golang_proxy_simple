package balancer

import "errors"

var (
	NoHostError                = errors.New("no host")
	AlgorithmNotSupportedError = errors.New("algorithm not supported")
)

// Balancer interface is the load balancer for the reverse proxy
type Balancer interface {
	Add(string)
	Remove(string)
	Balance(string) (string, error)
	Inc(string)
	Done(string)
}

var factories = make(map[string]Factory)

// Factory 是生成Balancer的工厂, 和工厂设计模式在这里使用
type Factory func([]string) Balancer

// Build 根据算法生成相应的负载均衡器
func Build(algorithm string, hosts []string) (Balancer, error) {
	factory, ok := factories[algorithm]
	if !ok {
		return nil, AlgorithmNotSupportedError
	}
	return factory(hosts), nil
}