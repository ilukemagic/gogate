package main

import (
	"flag"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ilukemagic/gogate/internal/config"
	"github.com/ilukemagic/gogate/internal/handler"
	"github.com/ilukemagic/gogate/internal/middleware"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 创建 JWT 中间件
	jwtMiddleware := middleware.NewJWTMiddleware(
		cfg.JWT.SecretKey,
		cfg.JWT.Exclude,
	)

	// 创建代理处理器
	proxyHandler, err := handler.NewProxyHandler(cfg.Proxy.Routes)
	if err != nil {
		log.Fatal("Failed to create proxy handler:", err)
	}

	// 创建 gin 引擎实例
	r := gin.Default()

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册登录路由
	r.POST("/api/auth/login", func(c *gin.Context) {
		var login struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&login); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		// todo: 这里简化了验证逻辑，实际应用中需要验证用户名和密码
		token, err := jwtMiddleware.GenerateToken("123", login.Username)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(200, gin.H{
			"token": token,
		})
	})

	// 注册API路由处理中间件
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 只处理/api开头且不是/api/auth开头的路径
		if strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/api/auth") {
			// 应用JWT中间件
			jwtMiddleware.Handle()(c)

			// 如果验证通过，继续处理
			if !c.IsAborted() {
				proxyHandler.Handle(c)
			}

			// 无论如何都不继续后续的处理器
			c.Abort()
		}
	})

	// 启动服务器
	log.Printf("Starting server on %s\n", cfg.Proxy.Listen)
	if err := r.Run(cfg.Proxy.Listen); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
