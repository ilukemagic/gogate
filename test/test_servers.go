package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func createServer(port string, serverID string) *gin.Engine {
	r := gin.Default()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("Hello from test server %s!", serverID),
			"server":  serverID,
		})
	})

	return r
}

func main() {
	// 创建两个测试服务器
	server1 := createServer(":8081", "1")
	server2 := createServer(":8082", "2")

	// 启动服务器
	go server1.Run(":8081")
	go server2.Run(":8082")

	// 防止主程序退出
	select {}
}
