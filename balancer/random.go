package balancer

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	factories[RandomBalancer] = NewRandom
}

//Random 随机算法：根据host长度生成随机数，动态获取其中某一个host进行转发
type Random struct {
	mux   sync.RWMutex
	hosts []string
	rnd   *rand.Rand
}

func NewRandom(hosts []string) Balancer {
	return &Random{
		hosts: hosts,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Random) Add(host string)  {
	r.mux.Lock()
	defer r.mux.Unlock()
	for _, h := range r.hosts {
		if h == host {
			return
		}
	}
	r.hosts = append(r.hosts, host)
}

func (r *Random) Remove(host string)  {
	r.mux.Lock()
	defer r.mux.Unlock()
	for i, h := range r.hosts {
		if h == host {
			r.hosts = append(r.hosts[:i], r.hosts[i+1:]...)
		}
	}
}

func (r *Random) Balance(string) (string,error) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	if len(r.hosts) == 0 {
		return "", NoHostError
	}
	return r.hosts[r.rnd.Intn(len(r.hosts))], nil
}

func (r *Random) Inc(string)  {

}

func (r *Random) Done(string)  {

}