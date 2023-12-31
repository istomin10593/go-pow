package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config - config for app.
type Config struct {
	Server struct {
		Host    string        `yaml:"host"`
		Port    string        `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"server"`
	Client struct {
		Timeout time.Duration `yaml:"timeout"`
		Number  int           `yaml:"number"`
	} `yaml:"client"`
	Pow struct {
		MaxIterations int `yaml:"maxiterations"`
	} `yaml:"pow"`
}

// Parse - parse config from file.
func Parse(filePath string) (*Config, error) {
	filename, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't get config path: %w", err)
	}

	yamlConf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("can't read conf: %w", err)
	}

	var config Config

	err = yaml.Unmarshal(yamlConf, &config)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshall conf: %w", err)
	}

	config.Server.Host = getEnv("SERVER_HOST", config.Server.Host)
	config.Server.Port = getEnv("SERVER_PORT", config.Server.Port)

	return &config, nil
}

// getEnv - get env value or default.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
