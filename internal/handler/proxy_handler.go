package handler

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ilukemagic/gogate/internal/proxy"
)

// ProxyHandler 处理代理请求
type ProxyHandler struct {
	proxies map[string]*proxy.ReverseProxy
}

// NewProxyHandler 创建新的代理处理器
func NewProxyHandler(routes map[string]string) (*ProxyHandler, error) {
	proxies := make(map[string]*proxy.ReverseProxy)

	for path, targetURL := range routes {
		p, err := proxy.NewReverseProxy(targetURL)
		if err != nil {
			return nil, err
		}
		proxies[path] = p
	}

	return &ProxyHandler{
		proxies: proxies,
	}, nil
}

// Handle 处理代理请求
func (h *ProxyHandler) Handle(c *gin.Context) {
	path := c.Request.URL.Path
	log.Printf("Received request for path: %s", path)

	var matchedProxy *proxy.ReverseProxy
	var longestMatch string

	for route, p := range h.proxies {
		log.Printf("Checking route: %s", route)
		if strings.HasPrefix(path, route) {
			if len(route) > len(longestMatch) {
				longestMatch = route
				matchedProxy = p
				log.Printf("Found matching route: %s", route)
			}
		}
	}

	if matchedProxy == nil {
		log.Printf("No matching route found for: %s", path)
		c.JSON(404, gin.H{"error": "route not found"})
		return
	}

	log.Printf("Proxying request to route: %s", longestMatch)
	matchedProxy.ServeHTTP(c.Writer, c.Request)
}
