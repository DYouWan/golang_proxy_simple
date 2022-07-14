package balancer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRandom_Add(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	cases := []struct {
		name   string
		lb     Balancer
		args   string
		expect Balancer
	}{
		{
			"test-1",
			&Random{hosts: []string{"http://127.0.0.1:1011", "http://127.0.0.1:1012", "http://127.0.0.1:1013"}, rnd: rnd},
			"http://127.0.0.1:1013",
			&Random{hosts: []string{"http://127.0.0.1:1011", "http://127.0.0.1:1012", "http://127.0.0.1:1013"}, rnd: rnd},
		},
		{
			"test-2",
			&Random{hosts: []string{"http://127.0.0.1:1011", "http://127.0.0.1:1012"}, rnd: rnd},
			"http://127.0.0.1:1012",
			&Random{hosts: []string{"http://127.0.0.1:1011", "http://127.0.0.1:1012"}, rnd: rnd},
		},
	}


	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.lb.Add(c.args)
			assert.Equal(t, c.expect, c.lb)
		})
	}
}

func BenchmarkRandom_Add(b *testing.B) {
	//生成随机数，随技术中包含重复的数据
	var old []string
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10000; i++ {
		r := rnd.Intn(100)
		old = append(old, strconv.Itoa(r))
	}
	//对随机数去重，结果值作为期望值
	var expected []string
	tempMap := map[string]byte{}
	for _, s := range old {
		l := len(tempMap)
		tempMap[s] = 0
		if len(tempMap) != l {
			expected = append(expected, s)
		}
	}
	sort.Slice(expected, func(i, j int) bool {
		e1, _ := strconv.Atoi(expected[i])
		e2, _ := strconv.Atoi(expected[j])
		return e1 < e2
	})

	//开启多个goroutine向hosts添加数据
	rd := Random{
		hosts: []string{},
		rnd:   rnd,
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < len(old); i++ {
				rd.Add(old[i])
			}
			wg.Done()
		}()
	}
	wg.Wait()
	sort.Slice(rd.hosts, func(i, j int) bool {
		e1, _ := strconv.Atoi(rd.hosts[i])
		e2, _ := strconv.Atoi(rd.hosts[j])
		return e1 < e2
	})
	fmt.Println(expected)
	fmt.Println(rd.hosts)
	assert.Equal(b, expected, rd.hosts)
}