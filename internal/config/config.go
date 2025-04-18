package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ProxyConfig 代理配置
type ProxyConfig struct {
	Listen string            `yaml:"listen"`
	Routes map[string]string `yaml:"routes"` // path -> targetURL 映射
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
			Routes: map[string]string{
				"/api/test": "http://localhost:8081",
			},
		},
	}
}
