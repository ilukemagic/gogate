package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ilukemagic/gogate/internal/config"
	"github.com/ilukemagic/gogate/internal/handler"
)

func main() {
	// 获取配置
	cfg := config.DefaultConfig()

	// 创建代理处理器
	proxyHandler, err := handler.NewProxyHandler(cfg.Proxy.Routes)
	if err != nil {
		log.Fatal("Failed to create proxy handler:", err)
	}

	// 创建 gin 引擎实例
	r := gin.Default()

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 注册代理路由
	r.Any("/api/*any", proxyHandler.Handle)

	// 启动服务器
	log.Printf("Starting server on %s\n", cfg.Proxy.Listen)
	if err := r.Run(cfg.Proxy.Listen); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
