package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ilukemagic/gogate/internal/balancer"
)

// 封装反向代理的基本功能
type ReverseProxy struct {
	balancer balancer.LoadBalancer
	proxies  map[string]*httputil.ReverseProxy
}

// 创建反向代理实例
func NewReverseProxy(targets []string) (*ReverseProxy, error) {
	// 创建负载均衡器
	lb := balancer.NewRoundRobin(targets)

	// 为每个目标创建代理
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, target := range targets {
		targetURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		proxies[target] = httputil.NewSingleHostReverseProxy(targetURL)
	}

	return &ReverseProxy{
		balancer: lb,
		proxies:  proxies,
	}, nil
}

// 实现 http.Handler 接口
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取下一个目标服务器
	target := p.balancer.Next()
	if target == "" {
		http.Error(w, "no available targets", http.StatusServiceUnavailable)
		return
	}

	// 获取对应的代理
	proxy := p.proxies[target]

	// 设置代理配置
	targetURL, _ := url.Parse(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host

		// 处理路径
		if strings.HasPrefix(req.URL.Path, "/api") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
		}

		// 设置代理相关的请求头
		req.Header.Set("X-Real-IP", req.RemoteAddr)
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Host = targetURL.Host
	}

	proxy.ServeHTTP(w, r)
}
