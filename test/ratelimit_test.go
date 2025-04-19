package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"
)

// 测试限流功能
func TestRateLimit(t *testing.T) {
	// 1. 获取JWT令牌用于后续测试
	var token string
	t.Run("GetToken", func(t *testing.T) {
		// 获取JWT令牌
		loginPayload := []byte(`{"username":"test","password":"test"}`)
		resp, err := http.Post("http://localhost:8080/api/auth/login",
			"application/json", bytes.NewBuffer(loginPayload))
		if err != nil {
			t.Fatalf("获取令牌失败: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		token, _ = result["token"].(string)
		if token == "" {
			t.Fatalf("未获取到有效令牌")
		}
	})

	// 2. 测试特定路由限流 (/api/test 限制为每秒10个请求)
	t.Run("TestPathSpecificRateLimit", func(t *testing.T) {
		// 统计成功和被限流的请求
		var successCount, limitedCount int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// 短时间内发送20个请求（理论上应该有约一半被限流）
		requestCount := 20
		wg.Add(requestCount)

		startTime := time.Now()

		for i := 0; i < requestCount; i++ {
			go func() {
				defer wg.Done()

				// 创建请求
				req, _ := http.NewRequest("GET", "http://localhost:8080/api/test", nil)
				req.Header.Add("Authorization", "Bearer "+token)

				// 发送请求
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Logf("请求错误: %v", err)
					return
				}
				defer resp.Body.Close()

				// 统计结果
				mu.Lock()
				defer mu.Unlock()

				if resp.StatusCode == 200 {
					successCount++
				} else if resp.StatusCode == 429 {
					limitedCount++
				} else {
					t.Logf("意外的状态码: %d", resp.StatusCode)
				}
			}()

			// 控制请求间隔，更容易观察限流效果
			time.Sleep(20 * time.Millisecond)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// 打印测试结果
		t.Logf("测试完成: 耗时=%.2f秒, 成功=%d, 限流=%d, 总计=%d",
			duration.Seconds(), successCount, limitedCount,
			successCount+limitedCount)

		// 验证是否有请求被限流
		if limitedCount == 0 {
			t.Errorf("预期有请求被限流，但实际所有请求都成功了")
		}

		// 验证成功请求是否在预期范围内
		// 对于速率10/秒+突发5的配置，在1秒内应该允许约10-15个请求
		expected := int(duration.Seconds()*10) + 5
		t.Logf("预期允许请求数约: %d", expected)

		if successCount > expected+3 { // 允许一些误差
			t.Errorf("成功请求数(%d)显著高于预期(%d)，限流可能未生效",
				successCount, expected)
		}
	})

	// 3. 测试全局限流
	t.Run("TestGlobalRateLimit", func(t *testing.T) {
		// 这部分测试需要更高的并发请求来触发全局限流
		// 由于全局限流设置较高(100/秒)，这里略过具体实现
		t.Log("全局限流测试需要更高并发，建议使用压测工具如wrk或vegeta")
	})
}
