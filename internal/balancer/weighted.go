package balancer

import "sync"

// 带权重的目标服务器
type WeightedTarget struct {
	URL           string
	Weight        int
	CurrentWeight int // 当前权重
}

// 权重轮询负载均衡器
type WeightedRoundRobin struct {
	targets []*WeightedTarget
	mu      sync.Mutex
}

// 创建权重轮询负载均衡器
func NewWeightedRoundRobin(targets map[string]int) *WeightedRoundRobin {
	wrr := &WeightedRoundRobin{
		targets: make([]*WeightedTarget, 0, len(targets)),
	}

	for url, weight := range targets {
		wrr.targets = append(wrr.targets, &WeightedTarget{
			URL:    url,
			Weight: weight,
		})
	}

	return wrr
}

// 获取下一个目标服务器(Nginx 平滑加权轮询算法)
func (w *WeightedRoundRobin) Next() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.targets) == 0 {
		return ""
	}

	totalWeight := 0

	var best *WeightedTarget

	// 为每个目标增加当前权重，并选择最大的一个
	for _, t := range w.targets {
		t.CurrentWeight += t.Weight
		totalWeight += t.Weight

		if best == nil || t.CurrentWeight > best.CurrentWeight {
			best = t
		}
	}

	// 选择的节点减去总权重
	if best != nil {
		best.CurrentWeight -= totalWeight
		return best.URL
	}

	return ""
}

// 更新目标服务器列表
func (w *WeightedRoundRobin) UpdateTargets(targets map[string]int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.targets = make([]*WeightedTarget, 0, len(targets))
	for url, weight := range targets {
		w.targets = append(w.targets, &WeightedTarget{
			URL:    url,
			Weight: weight,
		})
	}
}
