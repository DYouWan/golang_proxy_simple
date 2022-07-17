package balancer

import "sync"

/*
RoundRobin 轮询算法：根据host长度模运算，依次获取下一个host进行转发
优点：到后端应用的请求更加均匀，使得每个服务器基本均衡
缺点：随着后端应用服务器的增加，缓存的命中率为下降，这种方式不会因为热点问题导致其中某一台
服务器负载过重
 */
type RoundRobin struct {
	i     uint64
	hosts []string
	mux   sync.RWMutex
}

func init()  {
	factories[R2Balancer] = NewRoundRobin
}
func NewRoundRobin(hosts []string) Balancer {
	return &RoundRobin{
		i:    0,
		hosts: hosts,
	}
}

func (r *RoundRobin) Add(host string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, h := range r.hosts {
		if h == host {
			return
		}
	}
	r.hosts = append(r.hosts, host)
}


func (r *RoundRobin) Remove(host string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	for i, h := range r.hosts {
		if h == host {
			r.hosts = append(r.hosts[:i], r.hosts[i+1:]...)
			return
		}
	}
}

func (r *RoundRobin) Balance(_ string)(string,error) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if len(r.hosts) == 0 {
		return "", NoHostError
	}
	host := r.hosts[r.i%uint64(len(r.hosts))]
	r.i++
	return host, nil
}

func (r *RoundRobin) Inc(_ string)  {}

func (r *RoundRobin) Done(_ string)  {}