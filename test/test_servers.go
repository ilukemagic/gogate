package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	server1 := createServer(":8081", "1 (Weight: 3)")
	server2 := createServer(":8082", "2 (Weight: 2)")

	// 启动服务器
	go server1.Run(":8081")
	go server2.Run(":8082")

	// 添加测试统计
	go func() {
		counts := make(map[string]int)
		total := 0

		for {
			resp, err := http.Get("http://localhost:8080/api/test")
			if err != nil {
				fmt.Printf("Error making request: %v\n", err)
				time.Sleep(time.Second)
				continue
			}

			defer resp.Body.Close()

			// 解析响应
			var result struct {
				Server string `json:"server"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Printf("Error decoding response: %v\n", err)
				continue
			}

			// 统计请求分布
			counts[result.Server]++
			total++

			// 每10次请求打印一次统计信息
			if total%10 == 0 {
				fmt.Printf("\n=== Request Distribution (Total: %d) ===\n", total)
				for server, count := range counts {
					percentage := float64(count) / float64(total) * 100
					fmt.Printf("%s: %d requests (%.1f%%)\n", server, count, percentage)
				}
				fmt.Println("=====================================")
			}

			time.Sleep(time.Second)
		}
	}()

	// 防止主程序退出
	select {}
}
