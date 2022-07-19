package balancer

import (
	"hash/crc32"
	"sync"
)

//IPHash IP哈希算法：通过对请求的host进行hash运算后取模
type IPHash struct {
	hosts []string
	mux   sync.RWMutex
}

func init()  {
	factories[IPHashBalancer] = NewIPHash
}

func NewIPHash(hosts []string)Balancer {
	return &IPHash{
		hosts: hosts,
	}
}

func (h *IPHash) Add(host string) {
	h.mux.Lock()
	defer h.mux.Unlock()
	for _, item := range h.hosts {
		if item == host {
			return
		}
	}
	h.hosts = append(h.hosts, host)
}
func (h *IPHash) Remove(host string) {
	h.mux.Lock()
	defer h.mux.Unlock()
	for i, item := range h.hosts {
		if item == host {
			h.hosts = append(h.hosts[:i], h.hosts[i+1:]...)
			return
		}
	}
}
func (h *IPHash) Balance(ip string)(string,error) {
	h.mux.RLock()
	defer h.mux.RUnlock()
	if len(h.hosts) == 0 {
		return "", NoHostError
	}
	value := crc32.ChecksumIEEE([]byte(ip)) % uint32(len(h.hosts))
	return h.hosts[value], nil
}

func (h *IPHash) Inc(_ string) {}

func (h *IPHash) Done(_ string) {}