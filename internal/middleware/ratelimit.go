package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilukemagic/gogate/internal/config"
)

// 令牌桶限流器
type TokenBucket struct {
	mu         sync.Mutex
	rate       float64   // 令牌生成速率，每秒生成的令牌数
	burst      int       // 桶的容量，最大令牌数
	tokens     float64   // 当前令牌数
	lastUpdate time.Time // 上次更新时间
}

// 创建新的令牌桶限流器
func NewTokenBucket(rate int, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       float64(rate),
		burst:      burst,
		tokens:     float64(burst), // 初始填满令牌桶
		lastUpdate: time.Now(),
	}
}

// Allow 判断是否允许请求通过
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 计算从上次更新到现在经过的时间
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = now

	// 计算新增的令牌数
	newTokens := elapsed * tb.rate

	// 更新令牌数，但不超过桶的容量
	if tb.tokens+newTokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	} else {
		tb.tokens += newTokens
	}

	// 如果至少有1个令牌，则消耗一个令牌并允许请求
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	// 没有足够的令牌，请求被拒绝
	return false
}

// RateLimiter 限流中间件
type RateLimiter struct {
	globalLimiter *TokenBucket
	routeLimiters map[string]*TokenBucket
	config        config.RateLimitConfig
}

// NewRateLimiter 创建限流中间件
func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	globalLimiter := NewTokenBucket(cfg.Rate, cfg.Burst)
	routeLimiters := make(map[string]*TokenBucket)

	// 为每个指定路由创建限流器
	for route, routeCfg := range cfg.Routes {
		routeLimiters[route] = NewTokenBucket(routeCfg.Rate, routeCfg.Burst)
	}

	return &RateLimiter{
		globalLimiter: globalLimiter,
		routeLimiters: routeLimiters,
		config:        cfg,
	}
}

// Handle 限流中间件处理函数
func (rl *RateLimiter) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未启用限流，则直接通过
		if !rl.config.Enable {
			c.Next()
			return
		}

		// 获取请求路径
		path := c.Request.URL.Path

		// 优先检查特定路由的限流
		var matchedLimiter *TokenBucket
		var longestMatch string

		// 查找最长匹配的路由
		for route, limiter := range rl.routeLimiters {
			if strings.HasPrefix(path, route) && len(route) > len(longestMatch) {
				matchedLimiter = limiter
				longestMatch = route
			}
		}

		// 应用特定路由限流
		if matchedLimiter != nil {
			if !matchedLimiter.Allow() {
				c.JSON(429, gin.H{"error": "too many requests"})
				c.Abort()
				return
			}
		}

		// 应用全局限流
		if !rl.globalLimiter.Allow() {
			c.JSON(429, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
