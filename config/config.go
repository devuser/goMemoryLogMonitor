package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPPort    int `yaml:"http_port"`
	TCPPort     int `yaml:"tcp_port"`
	CacheSizeMB int `yaml:"cache_size_mb"`
}

func Default() Config {
	return Config{
		HTTPPort:    8080,
		TCPPort:     9090,
		CacheSizeMB: 100,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal yaml: %w", err)
	}
	return cfg, nil
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

