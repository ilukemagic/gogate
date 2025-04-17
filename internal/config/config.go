package config

// ProxyConfig 代理配置
type ProxyConfig struct {
	Listen string            `yaml:"listen"`
	Routes map[string]string `yaml:"routes"` // path -> targetURL 映射
}

// Config 全局配置
type Config struct {
	Proxy ProxyConfig `yaml:"proxy"`
}

// DefaultConfig 返回默认配置
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
