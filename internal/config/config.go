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

	// RAG configuration
	RAG RAGConfig `yaml:"rag"`

	// Runtime config
	ConfigPath string `yaml:"-"`
}

type RAGConfig struct {
	KnowledgeBase string   `yaml:"knowledge_base"`
	EmbedModel    string   `yaml:"embed_model"`
	ChunkSize     int      `yaml:"chunk_size"`
	ChunkOverlap  int      `yaml:"chunk_overlap"`
	AllowedFiles  []string `yaml:"allowed_files"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Host:     DefaultHost,
		Port:     DefaultPort,
		LogLevel: DefaultLogLevel,
		Verbose:  false,
		Quiet:    false,
		RAG: RAGConfig{
			EmbedModel:   "mxbai-embed-large",
			ChunkSize:    500,
			ChunkOverlap: 50,
		},
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

	// Generate config with inline comments
	configContent := c.generateConfigWithComments()

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) generateConfigWithComments() string {
	allowedFilesYAML := "[]"
	if len(c.RAG.AllowedFiles) > 0 {
		allowedFilesYAML = "\n"
		for _, file := range c.RAG.AllowedFiles {
			allowedFilesYAML += fmt.Sprintf("    - %s\n", file)
		}
		allowedFilesYAML += "  "
	}

	return fmt.Sprintf(`# Ollama Server Configuration
# The hostname or IP address of the Ollama server
host: %s

# The port number of the Ollama server (default: 11434)
port: %d

# Authentication token for the Ollama server (leave empty if not required)
token: "%s"

# Logging level: debug, info, warn, error (default: info)
log_level: %s

# Enable verbose output for debugging purposes
verbose: %t

# Enable quiet mode to suppress non-essential output
quiet: %t

# RAG (Retrieval Augmented Generation) Configuration
rag:
  # Path to the SQLite vector database for knowledge base storage
  # Leave empty to use default: ~/.ollamacli/knowledge.db
  knowledge_base: "%s"

  # Embedding model to use for vectorizing documents
  # Recommended models: mxbai-embed-large, nomic-embed-text
  embed_model: %s

  # Maximum size of each text chunk in characters (default: 500)
  # Larger chunks preserve more context but may reduce precision
  chunk_size: %d

  # Number of overlapping characters between chunks (default: 50)
  # Overlap helps maintain context across chunk boundaries
  chunk_overlap: %d

  # List of file patterns to restrict RAG queries (optional)
  # Example: ["*.go", "docs/*.md"]
  allowed_files: %s
`,
		c.Host,
		c.Port,
		c.Token,
		c.LogLevel,
		c.Verbose,
		c.Quiet,
		c.RAG.KnowledgeBase,
		c.RAG.EmbedModel,
		c.RAG.ChunkSize,
		c.RAG.ChunkOverlap,
		allowedFilesYAML,
	)
}

func (c *Config) GetServerURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

func (c *Config) HasToken() bool {
	return c.Token != ""
}

func (c *Config) GetKnowledgeBasePath() string {
	if c.RAG.KnowledgeBase != "" {
		return c.RAG.KnowledgeBase
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".ollamacli/knowledge.db"
	}

	return filepath.Join(homeDir, ".ollamacli", "knowledge.db")
}