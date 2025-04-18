package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// 封装反向代理的基本功能
type ReverseProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

// 创建一个新的反向代理实例
func NewReverseProxy(targetURL string) (*ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 设置代理的配置
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// 处理路径重写
		originalPath := req.URL.Path
		if strings.HasPrefix(originalPath, "/api") {
			// 如果目标URL有自己的路径，需要正确拼接
			if target.Path != "" {
				req.URL.Path = target.Path + strings.TrimPrefix(originalPath, "/api")
			} else {
				req.URL.Path = strings.TrimPrefix(originalPath, "/api")
			}
		}

		// 设置代理相关的请求头
		req.Header.Set("X-Real-IP", req.RemoteAddr)
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Host = target.Host
	}

	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "proxy error: "+err.Error(), http.StatusBadGateway)
	}

	return &ReverseProxy{
		target: target,
		proxy:  proxy,
	}, nil
}

// 实现 http.Handler 接口
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
