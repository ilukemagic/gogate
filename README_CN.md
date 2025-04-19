# GoGate - 轻量级云原生 API 网关

GoGate 是一个高性能、轻量级的 API 网关，专为云原生环境设计，注重简洁性和开发者体验。

## 功能特性

已实现的核心功能：

- **反向代理**：将请求转发到后端服务

  - 支持路径重写
  - 标准代理请求头处理（X-Forwarded-For, X-Real-IP 等）
  - 灵活的路由匹配规则

- **负载均衡**：智能分发请求到多个后端服务

  - 权重轮询算法（Weighted Round Robin）
  - 动态服务节点管理
  - 平滑的请求分配

- **JWT 鉴权**：保护 API 安全

  - 支持标准 JWT token 验证
  - 可配置的路径排除列表
  - 用户信息传递

- **限流控制**：防止服务过载
  - 令牌桶算法实现
  - 全局和路径级别限流
  - 可配置的速率和突发流量设置

## 安装与使用

### 前置条件

- Go 1.16+
- Git

### 安装步骤

```bash
# 克隆仓库
git clone https://github.com/ilukemagic/gogate.git
cd gogate

# 下载依赖
go mod download

# 编译
go build -o gogate cmd/server/main.go
```

### 配置

GoGate 使用 YAML 配置文件。在 `configs/config.yaml` 中修改配置：

```yaml
proxy:
  listen: ":8080" # 网关监听地址
  routes:
    "/api/test": # 路由路径
      targets:
        - url: "http://localhost:8081"
          weight: 3 # 权重为3
        - url: "http://localhost:8082"
          weight: 2 # 权重为2

jwt:
  secretKey: "your-secret-key-here"
  exclude: # 不需要JWT验证的路径
    - "/health"

rateLimit:
  enable: true
  rate: 100 # 全局限流：每秒100个请求
  burst: 50 # 允许突发50个请求
  routes:
    "/api/test":
      rate: 10 # 特定路径限流：每秒10个请求
      burst: 5 # 允许突发5个请求
```

### 运行

```bash
# 使用默认配置运行
./gogate

# 指定配置文件
./gogate -config path/to/config.yaml
```

## 测试

### 反向代理与负载均衡测试

1. 启动测试服务器：

```bash
go run test/test_servers.go
```

2. 启动 GoGate：

```bash
go run cmd/server/main.go
```

3. 获取 JWT 令牌：

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'
```

4. 测试 API 访问：

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/test
```

### 负载均衡测试

```bash
# 多次请求观察负载分布
for i in {1..10}; do
  curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/test
done
```

### 限流测试

```bash
# 使用测试脚本
./scripts/test_ratelimit.sh
```

## 贡献

欢迎贡献代码、报告问题或提出新功能建议。

## 许可证

本项目采用 MIT 许可证。

[English Version README](README.md)
