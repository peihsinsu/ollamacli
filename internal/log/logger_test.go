package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		verbose  bool
		expected Level
	}{
		{"debug level", "debug", false, LevelDebug},
		{"info level", "info", false, LevelInfo},
		{"warn level", "warn", false, LevelWarn},
		{"error level", "error", false, LevelError},
		{"unknown level defaults to info", "unknown", false, LevelInfo},
		{"verbose overrides to debug", "error", true, LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.verbose).(*logger)
			if logger.level != tt.expected {
				t.Errorf("Expected level %d, got %d", tt.expected, logger.level)
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer

	logger := New("debug", true).(*logger)
	logger.std.SetOutput(&stdout)
	logger.err.SetOutput(&stderr)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Check that debug, info, warn go to stdout
	if !strings.Contains(stdoutStr, "[DEBUG] debug message") {
		t.Error("Debug message not found in stdout")
	}
	if !strings.Contains(stdoutStr, "[INFO] info message") {
		t.Error("Info message not found in stdout")
	}
	if !strings.Contains(stdoutStr, "[WARN] warn message") {
		t.Error("Warn message not found in stdout")
	}

	// Check that error goes to stderr
	if !strings.Contains(stderrStr, "[ERROR] error message") {
		t.Error("Error message not found in stderr")
	}
}

func TestLoggerQuietMode(t *testing.T) {
	var stdout bytes.Buffer

	logger := New("info", false).(*logger)
	logger.std.SetOutput(&stdout)

	logger.Debug("debug message")
	logger.Info("info message")

	output := stdout.String()

	// In quiet mode (non-verbose), debug should not appear
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear in quiet mode")
	}

	// Info should appear but without level prefix
	if strings.Contains(output, "[INFO]") {
		t.Error("Level prefix should not appear in quiet mode")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message should appear in quiet mode")
	}
}

func TestSetLevel(t *testing.T) {
	logger := &logger{}

	logger.SetLevel("warn", false)
	if logger.level != LevelWarn {
		t.Errorf("Expected level %d, got %d", LevelWarn, logger.level)
	}

	logger.SetLevel("error", true)
	if logger.level != LevelDebug {
		t.Errorf("Verbose should override to debug level, got %d", logger.level)
	}
}