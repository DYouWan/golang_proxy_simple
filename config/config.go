package config

import (
	"errors"
	"fmt"
	"github.com/jinzhu/configor"
)

type Config struct {
	Server  Server    `yaml:"server"`
	Routing []Routing `yaml:"routing"`
}

type Server struct {
	Ip                  string `yaml:"ip" default:"127.0.0.1"`
	Port                int    `yaml:"port" default:"8080"`
	Schema              string `yaml:"schema" default:"http"`
	MaxAllowed          uint   `yaml:"max_allowed" default:"100"`
	CertKey             string `yaml:"cert_key"`
	CertCrt             string `yaml:"cert_crt"`
	HealthCheck         bool   `yaml:"tcp_health_check"`
	HealthCheckInterval uint   `yaml:"health_check_interval"`
}

type Routing struct {
	Pattern     string   `yaml:"pattern"`
	ProxyPass   []string `yaml:"proxy_pass"`
	BalanceMode string   `yaml:"balance_mode"`
}


func ReadConfig(configFile string,isValidation bool) (*Config, error) {
	if configFile == "" {
		return nil, fmt.Errorf("invalid file path")
	}
	c := &Config{}
	err := configor.Load(c, configFile)
	if err != nil {
		return nil, err
	}
	if isValidation {
		return c, c.Validation()
	}
	return c, nil
}

func (c *Config) Validation() error {
	if c.Server.Schema != "http" && c.Server.Schema != "https" {
		return fmt.Errorf("the schema \"%s\" not supported", c.Server.Schema)
	}
	if c.Server.Schema == "https" && (len(c.Server.CertCrt) == 0 || len(c.Server.CertKey) == 0) {
		return errors.New("the https proxy requires ssl_certificate_key and ssl_certificate")
	}
	if len(c.Routing) == 0 {
		return errors.New("the details of location cannot be null")
	}

	if c.Server.HealthCheckInterval < 1 {
		return errors.New("health_check_interval must be greater than 0")
	}
	return nil
}
