package balancer

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
)

//hostReplicaFormat 主机副本名称的格式
const hostReplicaFormat = `%s#%d`

var (
	//defaultReplicaNum 默认每个主机副本的数量
	defaultReplicaNum = 10

	//loadBoundFactor 负载因子
	loadBoundFactor = 0.25

	//defaultHashFunc 默认hash函数
	defaultHashFunc = func(key string) uint64 {
		out := sha512.Sum512([]byte(key))
		return binary.LittleEndian.Uint64(out[:])
	}
)

//ConsistentHash 是一个一致性哈希算法的实现
type ConsistentHash struct {
	//hashFunc 哈希函数
	hashFunc Hash

	//totalLoad 所有副本的总负载
	totalLoad int64

	//replicaNum 每个主机副本的数量
	replicaNum int

	// 哈希环
	sortedHostsHashSet []uint64

	//replicaHostMap 散列虚拟节点到主机名的映射
	replicaHostMap map[uint64]string

	//hostMap 虚拟节点到主机的映射
	hostMap map[string]*ConsistentHashHost

	sync.RWMutex
}

type ConsistentHashHost struct {
	// the host id: ip:port
	Name string
	// the load bound of the host
	LoadBound int64
}

func init()  {
	factories[ConsistentHashBalancer] = NewConsistentHash
}

func NewConsistentHash(hosts []string) Balancer {
	ch := &ConsistentHash{
		replicaNum:         defaultReplicaNum,
		totalLoad:          0,
		hashFunc:           defaultHashFunc,
		hostMap:            make(map[string]*ConsistentHashHost),
		replicaHostMap:     make(map[uint64]string),
		sortedHostsHashSet: make([]uint64, 0),
	}
	for _, host := range hosts {
		ch.Add(host)
	}
	return ch
}

func NewConsistent(replicaNum int, hash Hash) *ConsistentHash {
	if replicaNum <= 0 {
		replicaNum = defaultReplicaNum
	}
	if hash == nil {
		hash = defaultHashFunc
	}
	return &ConsistentHash{
		replicaNum:         replicaNum,
		totalLoad:          0,
		hashFunc:           hash,
		hostMap:            make(map[string]*ConsistentHashHost),
		replicaHostMap:     make(map[uint64]string),
		sortedHostsHashSet: make([]uint64, 0),
	}
}

func (c *ConsistentHash) Configuration(replicaNum int,fn Hash) {
	if replicaNum <= 0 {
		replicaNum = defaultReplicaNum
	}
	if fn == nil {
		fn = defaultHashFunc
	}
	c.hashFunc = fn
	c.replicaNum = replicaNum
}

//Add 添加主机
func (c *ConsistentHash) Add(hostName string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.hostMap[hostName]; ok {
		return
	}

	c.hostMap[hostName] = &ConsistentHashHost{
		Name:      hostName,
		LoadBound: 0,
	}

	for i := 0; i < c.replicaNum; i++ {
		hashedIdx := c.hashFunc(fmt.Sprintf(hostReplicaFormat, hostName, i))
		c.replicaHostMap[hashedIdx] = hostName
		c.sortedHostsHashSet = append(c.sortedHostsHashSet, hashedIdx)
	}

	// 按升序排序哈希值
	sort.Slice(c.sortedHostsHashSet, func(i int, j int) bool {
		if c.sortedHostsHashSet[i] < c.sortedHostsHashSet[j] {
			return true
		}
		return false
	})
}

//Remove 删除主机及主机的副本
func (c *ConsistentHash) Remove(hostName string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.hostMap[hostName]; !ok {
		return
	}

	delete(c.hostMap, hostName)

	for i := 0; i < c.replicaNum; i++ {
		hashedIdx := c.hashFunc(fmt.Sprintf(hostReplicaFormat, hostName, i))
		delete(c.replicaHostMap, hashedIdx)
		c.delHashIndex(hashedIdx)
	}
}

//Balance 通过key获取目标主机
func (c *ConsistentHash) Balance(key string) (string, error) {
	if len(c.hostMap) == 0 {
		return "", NoHostError
	}
	hashedKey := c.hashFunc(key)
	idx := c.searchKey(hashedKey)
	return c.replicaHostMap[c.sortedHostsHashSet[idx]], nil
}

//Inc 主机负载增加1 应该只在通过GetLeast获取主机时使用
func (c *ConsistentHash) Inc(hostName string) {
	c.Lock()
	defer c.Unlock()

	atomic.AddInt64(&c.hostMap[hostName].LoadBound, 1)
	atomic.AddInt64(&c.totalLoad, 1)
}

//Done 将主机负载减1 应该只在通过GetLeast获取主机时使用
func (c *ConsistentHash) Done(host string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.hostMap[host]; !ok {
		return
	}
	atomic.AddInt64(&c.hostMap[host].LoadBound, -1)
	atomic.AddInt64(&c.totalLoad, -1)
}

// Hosts 返回真实主机列表
func (c *ConsistentHash) Hosts() []string {
	c.RLock()
	defer c.RUnlock()

	hosts := make([]string, 0)
	for k := range c.hostMap {
		hosts = append(hosts, k)
	}
	return hosts
}

// UpdateLoad 设置host对应的LoadBound
func (c *ConsistentHash) UpdateLoad(host string, load int64) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.hostMap[host]; !ok {
		return
	}
	c.totalLoad = c.totalLoad - c.hostMap[host].LoadBound + load
	c.hostMap[host].LoadBound = load
}

//GetKeyLeast 它使用有界负载的一致哈希选择最小负载的主机来提供密钥。如果环中没有主机，则返回ErrNoHosts。
func (c *ConsistentHash) GetKeyLeast(key string) (string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.replicaHostMap) == 0 {
		return "", ErrHostNotFound
	}

	hashedKey := c.hashFunc(key)
	idx := c.searchKey(hashedKey) // Find the first host that may serve the key

	i := idx
	for {
		host := c.replicaHostMap[c.sortedHostsHashSet[i]]
		loadChecked, err := c.checkLoadCapacity(host)
		if err != nil {
			return "", err
		}
		if loadChecked {
			return host, nil
		}
		i++

		// if idx goes to the end of the ring, start from the beginning
		if i >= len(c.replicaHostMap) {
			i = 0
		}
	}
}

// GetLoads 返回所有主机的负载
func (c *ConsistentHash) GetLoads() map[string]int64 {
	c.RLock()
	defer c.RUnlock()

	loads := make(map[string]int64)
	for k, v := range c.hostMap {
		loads[k] = atomic.LoadInt64(&v.LoadBound)
	}
	return loads
}

//MaxLoad 返回单个主机的最大负载
//(total_load / number_of_hosts) * 1.25
//total_load是主机服务的活动请求的总数
func (c *ConsistentHash) MaxLoad() int64 {
	if c.totalLoad == 0 {
		c.totalLoad = 1
	}

	var avgLoadPerNode float64
	avgLoadPerNode = float64(c.totalLoad / int64(len(c.hostMap)))
	if avgLoadPerNode == 0 {
		avgLoadPerNode = 1
	}
	avgLoadPerNode = math.Ceil(avgLoadPerNode * (1 + loadBoundFactor))
	return int64(avgLoadPerNode)
}

//delHashIndex 从散列环中移除散列主机索引
func (c *ConsistentHash) delHashIndex(val uint64) {
	idx := -1
	l := 0
	r := len(c.sortedHostsHashSet) - 1
	for l <= r {
		m := (l + r) / 2
		if c.sortedHostsHashSet[m] == val {
			idx = m
			break
		} else if c.sortedHostsHashSet[m] < val {
			l = m + 1
		} else if c.sortedHostsHashSet[m] > val {
			r = m - 1
		}
	}
	if idx != -1 {
		c.sortedHostsHashSet = append(c.sortedHostsHashSet[:idx], c.sortedHostsHashSet[idx+1:]...)
	}
}

//searchKey 从散列环中获取主机索引
func (c *ConsistentHash) searchKey(key uint64) int {
	idx := sort.Search(len(c.sortedHostsHashSet), func(i int) bool {
		return c.sortedHostsHashSet[i] >= key
	})

	if idx >= len(c.sortedHostsHashSet) {
		// make search as a ring
		idx = 0
	}

	return idx
}

//checkLoadCapacity 检查主机是否可以在负载范围内提供密钥
func (c *ConsistentHash) checkLoadCapacity(host string) (bool, error) {

	// a safety check if someone performed c.Done more than needed
	if c.totalLoad < 0 {
		c.totalLoad = 0
	}

	var avgLoadPerNode float64
	avgLoadPerNode = float64((c.totalLoad + 1) / int64(len(c.hostMap)))
	if avgLoadPerNode == 0 {
		avgLoadPerNode = 1
	}
	avgLoadPerNode = math.Ceil(avgLoadPerNode * (1 + loadBoundFactor))

	candidateHost, ok := c.hostMap[host]
	if !ok {
		return false, ErrHostNotFound
	}

	if float64(candidateHost.LoadBound)+1 <= avgLoadPerNode {
		return true, nil
	}

	return false, nil
}


