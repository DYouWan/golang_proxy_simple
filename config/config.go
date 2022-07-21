package config

import (
	"errors"
	"fmt"
	"github.com/jinzhu/configor"
)

const Algorithms string = "ip-hash|consistent-hash|p2c|random|round-robin|least-load|bounded"

type Config struct {
	Port                int       `yaml:"port" default:"8080"`
	Schema              string    `yaml:"schema" default:"http"`
	MaxAllowed          uint      `yaml:"max_allowed" default:"100"`
	CertKey             string    `yaml:"cert_key"`
	CertCrt             string    `yaml:"cert_crt"`
	HealthCheck         bool      `yaml:"health_check"`
	HealthCheckInterval uint      `yaml:"health_check_interval"`
	Routes              []Routing `json:"ReRoutes"`
}

func Read(isValidation bool,files ...string) (*Config, error) {
	if files == nil || len(files) == 0 {
		return nil, fmt.Errorf("无效的配置文件路径")
	}
	cfg := &Config{}
	err := configor.Load(cfg, files...)
	if err != nil {
		return nil, err
	}
	if isValidation {
		return cfg, cfg.Validation()
	}
	return cfg, nil
}

func (c *Config) Validation() error {
	if c.Schema != "http" && c.Schema != "https" {
		return fmt.Errorf("\"%s\" 模式不正确", c.Schema)
	}
	if c.Schema == "https" && (len(c.CertCrt) == 0 || len(c.CertKey) == 0) {
		return errors.New("HTTPS代理需要ssl_certificate_key和ssl_certificate")
	}
	if len(c.Routes) == 0 {
		return errors.New("路由配置不正确，至少要配置一个路由")
	}
	if c.HealthCheckInterval < 1 {
		return errors.New("健康检查间隔时间必须大于0")
	}
	return nil
}
