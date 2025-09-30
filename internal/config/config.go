package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	DefaultHost     = "localhost"
	DefaultPort     = 11434
	DefaultLogLevel = "info"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Token    string `yaml:"token"`
	LogLevel string `yaml:"log_level"`
	Verbose  bool   `yaml:"verbose"`
	Quiet    bool   `yaml:"quiet"`

	// Runtime config
	ConfigPath string `yaml:"-"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Host:     DefaultHost,
		Port:     DefaultPort,
		LogLevel: DefaultLogLevel,
		Verbose:  false,
		Quiet:    false,
	}

	// Load from config file
	if err := cfg.loadFromFile(); err != nil {
		// Config file is optional, so we only return error if file exists but can't be read
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg, nil
}

func (c *Config) loadFromFile() error {
	configPath := c.getConfigPath()
	c.ConfigPath = configPath

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

func (c *Config) loadFromEnv() {
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		c.Host = host
	}

	if portStr := os.Getenv("OLLAMA_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			c.Port = port
		}
	}

	if token := os.Getenv("OLLAMA_TOKEN"); token != "" {
		c.Token = token
	}

	if logLevel := os.Getenv("OLLAMA_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}

	if verboseStr := os.Getenv("OLLAMA_VERBOSE"); verboseStr != "" {
		c.Verbose = verboseStr == "true" || verboseStr == "1"
	}

	if quietStr := os.Getenv("OLLAMA_QUIET"); quietStr != "" {
		c.Quiet = quietStr == "true" || quietStr == "1"
	}
}

func (c *Config) getConfigPath() string {
	if configPath := os.Getenv("OLLAMA_CONFIG_PATH"); configPath != "" {
		return configPath
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".ollamacli/config.yaml"
	}

	return filepath.Join(homeDir, ".ollamacli", "config.yaml")
}

func (c *Config) Save() error {
	configPath := c.getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) GetServerURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

func (c *Config) HasToken() bool {
	return c.Token != ""
}