package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseProxy 封装反向代理的基本功能
type ReverseProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

// NewReverseProxy 创建一个新的反向代理实例
func NewReverseProxy(targetURL string) (*ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义 Director 以处理请求
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	return &ReverseProxy{
		target: target,
		proxy:  proxy,
	}, nil
}

// ServeHTTP 实现 http.Handler 接口
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
