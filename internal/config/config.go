package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// RouteConfig 路由配置
type RouteConfig struct {
	Targets []string `yaml:"targets"` // 支持多个目标服务器
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Listen string                 `yaml:"listen"`
	Routes map[string]RouteConfig `yaml:"routes"`
}

// Config 全局配置
type Config struct {
	Proxy ProxyConfig `yaml:"proxy"`
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

// DefaultConfig 仅用于开发测试
func DefaultConfig() *Config {
	return &Config{
		Proxy: ProxyConfig{
			Listen: ":8080",
			Routes: map[string]RouteConfig{
				"/api/test": {Targets: []string{"http://localhost:8081"}},
			},
		},
	}
}
