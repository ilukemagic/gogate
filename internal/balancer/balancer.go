package balancer

import (
	"sync/atomic"
)

// 负载均衡器接口
type LoadBalancer interface {
	Next() string
	UpdateTargets(targets []string)
}

// 轮询负载均衡
type RoundRobin struct {
	targets []string
	current uint64
}

// 实例化负载均衡器
func NewRoundRobin(targets []string) *RoundRobin {
	return &RoundRobin{
		targets: targets,
		current: 0,
	}
}

// 获取下一个目标服务器
func (r *RoundRobin) Next() string {
	if len(r.targets) == 0 {
		return ""
	}

	// 原子操作确保并发安全
	current := atomic.AddUint64(&r.current, 1)
	return r.targets[current%uint64(len(r.targets))]
}

// 更新目标服务器列表
func (r *RoundRobin) UpdateTargets(targets []string) {
	r.targets = targets
}
