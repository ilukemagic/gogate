package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	// 跳过集成测试，除非明确指定
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 查找项目根目录
	rootDir, err := findProjectRoot()
	if err != nil {
		t.Fatalf("无法找到项目根目录: %v", err)
	}

	// 确保配置文件存在
	configPath := filepath.Join(rootDir, "configs", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("配置文件不存在: %s", configPath)
	}

	// 启动测试服务器
	serverCmd := exec.Command("go", "run", filepath.Join(rootDir, "test", "test_servers.go"))
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Dir = rootDir // 设置工作目录
	if err := serverCmd.Start(); err != nil {
		t.Fatalf("无法启动测试服务器: %v", err)
	}
	defer serverCmd.Process.Kill()

	// 启动网关服务
	gatewayCmd := exec.Command("go", "run", filepath.Join(rootDir, "cmd", "server", "main.go"))
	gatewayCmd.Stdout = os.Stdout
	gatewayCmd.Stderr = os.Stderr
	gatewayCmd.Dir = rootDir // 设置工作目录
	if err := gatewayCmd.Start(); err != nil {
		t.Fatalf("无法启动网关服务: %v", err)
	}
	defer gatewayCmd.Process.Kill()

	// 等待服务启动
	waitForServer(t, "http://localhost:8080/health", 10)

	// 1. 测试健康检查
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("期望状态码 200，获得 %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if result["status"] != "ok" {
			t.Fatalf("期望状态为 'ok'，获得 %v", result["status"])
		}
	})

	// 2. 测试未授权访问
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/api/test")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 401 {
			t.Fatalf("期望状态码 401，获得 %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if result["error"] != "authorization header is required" {
			t.Fatalf("期望错误消息，获得 %v", result["error"])
		}
	})

	// 3. 获取JWT令牌
	var token string
	t.Run("GetJWTToken", func(t *testing.T) {
		loginPayload := []byte(`{"username":"test","password":"test"}`)
		resp, err := http.Post("http://localhost:8080/api/auth/login",
			"application/json", bytes.NewBuffer(loginPayload))
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("期望状态码 200，获得 %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}

		if result["token"] == nil {
			t.Fatalf("响应中没有token字段")
		}

		token = result["token"].(string)
	})

	// 4. 测试负载均衡和反向代理
	t.Run("LoadBalancingAndProxy", func(t *testing.T) {
		if token == "" {
			t.Fatal("测试依赖于有效的JWT令牌")
		}

		// 计数器
		serverCounts := make(map[string]int)
		var mu sync.Mutex
		var wg sync.WaitGroup

		// 发送多个并发请求
		totalRequests := 30
		wg.Add(totalRequests)

		for i := 0; i < totalRequests; i++ {
			go func() {
				defer wg.Done()

				client := &http.Client{}
				req, _ := http.NewRequest("GET", "http://localhost:8080/api/test", nil)
				req.Header.Add("Authorization", "Bearer "+token)

				resp, err := client.Do(req)
				if err != nil {
					t.Logf("请求失败: %v", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					t.Logf("请求返回非200状态码: %d", resp.StatusCode)
					return
				}

				body, _ := io.ReadAll(resp.Body)
				responseStr := string(body)

				mu.Lock()
				defer mu.Unlock()

				// 统计服务器分布
				if strings.Contains(responseStr, "1 (Weight: 3)") {
					serverCounts["server1"]++
				} else if strings.Contains(responseStr, "2 (Weight: 2)") {
					serverCounts["server2"]++
				}
			}()
		}

		wg.Wait()

		// 验证负载均衡
		server1Count := serverCounts["server1"]
		server2Count := serverCounts["server2"]
		total := server1Count + server2Count

		if total == 0 {
			t.Fatal("没有收到任何有效响应")
		}

		server1Ratio := float64(server1Count) / float64(total)
		server2Ratio := float64(server2Count) / float64(total)

		t.Logf("服务器分布: 服务器1=%d (%.1f%%), 服务器2=%d (%.1f%%)",
			server1Count, server1Ratio*100,
			server2Count, server2Ratio*100)

		// 检查服务器1是否获得更多请求（权重为3）
		if server1Count <= server2Count && total >= 10 {
			t.Logf("警告: 服务器1 (权重3) 应该接收比服务器2 (权重2) 更多的请求")
		}
	})
}

// 查找项目根目录
func findProjectRoot() (string, error) {
	// 先尝试当前目录
	if _, err := os.Stat("configs/config.yaml"); err == nil {
		absPath, err := filepath.Abs(".")
		return absPath, err
	}

	// 向上查找
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "configs", "config.yaml")); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return "", fmt.Errorf("无法找到项目根目录")
}

// 等待服务器启动
func waitForServer(t *testing.T, url string, timeoutSeconds int) {
	t.Logf("等待服务器启动: %s", url)

	for i := 0; i < timeoutSeconds; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				t.Logf("服务器已启动")
				return
			}
		}
		time.Sleep(time.Second)
	}

	t.Logf("警告: 等待服务器启动超时，测试可能会失败")
}
