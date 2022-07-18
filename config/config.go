package config

import (
	"errors"
	"fmt"
	"github.com/jinzhu/configor"
)

type Config struct {
	Port                int       `yaml:"port" default:"8080"`
	Schema              string    `yaml:"schema" default:"http"`
	MaxAllowed          uint      `yaml:"max_allowed" default:"100"`
	CertKey             string    `yaml:"cert_key"`
	CertCrt             string    `yaml:"cert_crt"`
	HealthCheck         bool      `yaml:"health_check"`
	HealthCheckInterval uint      `yaml:"health_check_interval"`
	Algorithms          []string  `yaml:"algorithms"`
	Routes              []Routing `json:"ReRoutes"`
}

func Read(isValidation bool,files ...string) (*Config, error) {
	if files == nil || len(files) == 0 {
		return nil, fmt.Errorf("invalid file path")
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
		return fmt.Errorf("the schema \"%s\" not supported", c.Schema)
	}
	if c.Schema == "https" && (len(c.CertCrt) == 0 || len(c.CertKey) == 0) {
		return errors.New("the https proxy requires ssl_certificate_key and ssl_certificate")
	}
	if len(c.Routes) == 0 {
		return errors.New("the details of location cannot be null")
	}
	if c.HealthCheckInterval < 1 {
		return errors.New("health_check_interval must be greater than 0")
	}
	return nil
}

func (c *Config) ValidationAlgorithm(str string) error {
	var exists bool
	for _, algorithm := range c.Algorithms {
		if algorithm == str {
			exists = true
		}
	}
	if exists == false {
		return fmt.Errorf("the algorithm \"%s\" not supported", str)
	}
	return nil
}

