package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// 路由配置
type RouteConfig struct {
	Targets []TargetConfig `yaml:"targets"` // 支持多个目标服务器
}

// 目标服务器配置
type TargetConfig struct {
	URL    string `yaml:"url"`    // 服务器地址
	Weight int    `yaml:"weight"` // 权重
}

// 代理配置
type ProxyConfig struct {
	Listen string                 `yaml:"listen"`
	Routes map[string]RouteConfig `yaml:"routes"`
}

// JWT 配置
type JWTConfig struct {
	SecretKey string   `yaml:"secretKey"`
	Exclude   []string `yaml:"exclude"` // 不需要验证的路径
}

// 全局配置
type Config struct {
	Proxy ProxyConfig `yaml:"proxy"`
	JWT   JWTConfig   `yaml:"jwt"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
