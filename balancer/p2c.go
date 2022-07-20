package balancer

import (
	"hash/crc32"
	"math/rand"
	"sync"
	"time"
)

func init() {
	factories[P2CBalancer] = NewP2C
}

const Salt = "%#!"

// P2C (Pick Of 2 Choices)首先随机选取两个节点，在这两个节点中选择延迟低，或者连接数小的节点处理请求，这样兼顾了随机性，又兼顾了机器的性能
type P2C struct {
	mux sync.RWMutex
	hosts   []*HostLoad
	rnd     *rand.Rand
	loadMap map[string]*HostLoad
}

// NewP2C create new P2C balancer
func NewP2C(hosts []string) Balancer {
	p := &P2C{
		hosts:   []*HostLoad{},
		loadMap: make(map[string]*HostLoad),
		rnd:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for _, h := range hosts {
		p.Add(h)
	}
	return p
}

// Add new host to the balancer
func (p *P2C) Add(hostName string) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if _, ok := p.loadMap[hostName]; ok {
		return
	}

	h := &HostLoad{name: hostName, load: 0}
	p.hosts = append(p.hosts, h)
	p.loadMap[hostName] = h
}

// Remove new host from the balancer
func (p *P2C) Remove(host string) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if _, ok := p.loadMap[host]; !ok {
		return
	}

	delete(p.loadMap, host)

	for i, h := range p.hosts {
		if h.name == host {
			p.hosts = append(p.hosts[:i], p.hosts[i+1:]...)
			return
		}
	}
}

// Inc refers to the number of connections to the server `+1`
func (p *P2C) Inc(host string) {
	p.mux.Lock()
	defer p.mux.Unlock()

	h, ok := p.loadMap[host]

	if !ok {
		return
	}
	h.load++
}

// Done refers to the number of connections to the server `-1`
func (p *P2C) Done(host string) {
	p.mux.Lock()
	defer p.mux.Unlock()

	h, ok := p.loadMap[host]

	if !ok {
		return
	}

	if h.load > 0 {
		h.load--
	}
}

// Balance selects a suitable host according to the key value
func (p *P2C) Balance(key string) (string, error) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	if len(p.hosts) == 0 {
		return "", NoHostError
	}

	n1, n2 := p.hash(key)
	host := n2
	if p.loadMap[n1].load <= p.loadMap[n2].load {
		host = n1
	}
	return host, nil
}

func (p *P2C) hash(key string) (string, string) {
	var n1, n2 string
	if len(key) > 0 {
		saltKey := key + Salt
		n1 = p.hosts[crc32.ChecksumIEEE([]byte(key))%uint32(len(p.hosts))].name
		n2 = p.hosts[crc32.ChecksumIEEE([]byte(saltKey))%uint32(len(p.hosts))].name
		return n1, n2
	}
	n1 = p.hosts[p.rnd.Intn(len(p.hosts))].name
	n2 = p.hosts[p.rnd.Intn(len(p.hosts))].name
	return n1, n2
}