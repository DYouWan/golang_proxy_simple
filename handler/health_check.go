package handler

import (
	"proxy/util"
	"proxy/util/logging"
	"time"
)

//HealthCheck 主机健康检查
func (rh *RoutePrefixHandler) HealthCheck(interval uint) {
	go func() {
		for host := range rh.reverseProxyMap {
			rh.healthCheck(host, interval)
		}
	}()
}

//healthCheck 主机健康检查
func (rh *RoutePrefixHandler) healthCheck(host string, interval uint) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		isBackendAlive := util.IsBackendAlive(host)
		if !isBackendAlive && rh.ReadAlive(host) {
			logging.Errorf("连接主机 %s 失败, 已将状态置为不可用", host)

			rh.SetAlive(host, false)
			rh.bl.Remove(host)
		} else if isBackendAlive && !rh.ReadAlive(host) {
			logging.Infof("连接主机 %s 成功, 已将状态置为存活", host)

			rh.SetAlive(host, true)
			rh.bl.Add(host)
		}
	}
}

// ReadAlive 获取主机存活状态
func (rh *RoutePrefixHandler) ReadAlive(url string) bool {
	rh.mux.RLock()
	defer rh.mux.RUnlock()
	return rh.alive[url]
}

// SetAlive 设置主机存活状态
func (rh *RoutePrefixHandler) SetAlive(url string, alive bool) {
	rh.mux.Lock()
	defer rh.mux.Unlock()
	rh.alive[url] = alive
}