package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	// Test default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Host != DefaultHost {
		t.Errorf("Expected host %s, got %s", DefaultHost, cfg.Host)
	}

	if cfg.Port != DefaultPort {
		t.Errorf("Expected port %d, got %d", DefaultPort, cfg.Port)
	}

	if cfg.LogLevel != DefaultLogLevel {
		t.Errorf("Expected log level %s, got %s", DefaultLogLevel, cfg.LogLevel)
	}
}

func TestConfigLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("OLLAMA_HOST", "test-host")
	os.Setenv("OLLAMA_PORT", "9999")
	os.Setenv("OLLAMA_TOKEN", "test-token")
	os.Setenv("OLLAMA_VERBOSE", "true")

	defer func() {
		os.Unsetenv("OLLAMA_HOST")
		os.Unsetenv("OLLAMA_PORT")
		os.Unsetenv("OLLAMA_TOKEN")
		os.Unsetenv("OLLAMA_VERBOSE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.Host != "test-host" {
		t.Errorf("Expected host test-host, got %s", cfg.Host)
	}

	if cfg.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", cfg.Port)
	}

	if cfg.Token != "test-token" {
		t.Errorf("Expected token test-token, got %s", cfg.Token)
	}

	if !cfg.Verbose {
		t.Errorf("Expected verbose true, got %v", cfg.Verbose)
	}
}

func TestConfigGetServerURL(t *testing.T) {
	cfg := &Config{
		Host: "example.com",
		Port: 8080,
	}

	expected := "http://example.com:8080"
	if url := cfg.GetServerURL(); url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}

func TestConfigHasToken(t *testing.T) {
	cfg := &Config{}

	if cfg.HasToken() {
		t.Error("Expected HasToken to be false for empty token")
	}

	cfg.Token = "some-token"
	if !cfg.HasToken() {
		t.Error("Expected HasToken to be true for non-empty token")
	}
}

func TestConfigSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Override the config path environment variable for testing
	os.Setenv("OLLAMA_CONFIG_PATH", configPath)
	defer os.Unsetenv("OLLAMA_CONFIG_PATH")

	cfg := &Config{
		Host:     "test-host",
		Port:     9999,
		Token:    "test-token",
		LogLevel: "debug",
		Verbose:  true,
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Expected no error saving config, got: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created: %s", configPath)
	}

	// Verify content can be loaded
	cfg2 := &Config{}
	if err := cfg2.loadFromFile(); err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}

	if cfg2.Host != cfg.Host {
		t.Errorf("Expected host %s, got %s", cfg.Host, cfg2.Host)
	}
}