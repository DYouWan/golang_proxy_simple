package balancer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoundRobin_Balance(t *testing.T) {
	hosts := []string{
		"http://localhost:8011/api1",
		"http://localhost:8011/api2",
		"http://localhost:8011/api3",
		"http://localhost:8011/api4",
		"http://localhost:8011/api5",
		"http://localhost:8011/api6"}
	roundRobin := NewRoundRobin(hosts)

	hostMap :=map[string]int{
		"http://localhost:8011/api1": 0,
		"http://localhost:8011/api2": 1,
		"http://localhost:8011/api3": 2,
		"http://localhost:8011/api4": 3,
		"http://localhost:8011/api5": 4,
		"http://localhost:8011/api6": 5,
	}
	for i := 0; i < 10000; i++ {
		host, _ := roundRobin.Balance("")
		index := hostMap[host]
		assert.Equal(t, index, i%6)
	}
}
