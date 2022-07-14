package balancer

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	factories[RandomBalancer] = NewRandom
}

type Random struct {
	rw sync.RWMutex
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
	r.rw.Lock()
	defer r.rw.Unlock()
	for _, h := range r.hosts {
		if h == host {
			return
		}
	}
	r.hosts = append(r.hosts, host)
}

func (r *Random) Remove(host string)  {
	r.rw.Lock()
	defer r.rw.Unlock()
	for i, h := range r.hosts {
		if h == host {
			r.hosts = append(r.hosts[:i], r.hosts[i+1:]...)
		}
	}
}

func (r *Random) Balance(string) (string,error) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if len(r.hosts) == 0 {
		return "", NoHostError
	}
	return r.hosts[r.rnd.Intn(len(r.hosts))], nil
}

func (r *Random) Inc(string)  {

}

func (r *Random) Done(string)  {

}