package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 gin 引擎实例
	r := gin.Default()

	//  设置健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
