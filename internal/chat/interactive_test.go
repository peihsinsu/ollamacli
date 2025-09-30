package chat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ollamacli/internal/client"
	"ollamacli/internal/log"
	"ollamacli/internal/output"
)

func TestNewInteractiveChat(t *testing.T) {
	var outputBuf strings.Builder
	inputBuf := strings.NewReader("test input")

	logger := log.New("info", false)
	formatter := output.New(output.Options{
		Format: output.FormatText,
		Writer: &outputBuf,
	})

	opts := Options{
		Formatter: formatter,
		Logger:    logger,
		Model:     "test-model",
		Writer:    &outputBuf,
		Reader:    inputBuf,
	}

	ic := NewInteractiveChat(opts)

	if ic.model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", ic.model)
	}

	if len(ic.messages) != 0 {
		t.Errorf("Expected empty message history, got %d messages", len(ic.messages))
	}

	if ic.prompt != DefaultPrompt {
		t.Errorf("Expected default prompt '%s', got '%s'", DefaultPrompt, ic.prompt)
	}
}

func TestInteractiveChatHistory(t *testing.T) {
	ic := &InteractiveChat{
		messages: make([]client.ChatMessage, 0),
	}

	// Test initial empty history
	history := ic.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d messages", len(history))
	}

	// Add messages
	testMessages := []client.ChatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	ic.SetHistory(testMessages)

	// Get history and verify
	history = ic.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 messages in history, got %d", len(history))
	}

	if history[0].Role != "user" || history[0].Content != "Hello" {
		t.Errorf("First message incorrect: %+v", history[0])
	}

	if history[1].Role != "assistant" || history[1].Content != "Hi there!" {
		t.Errorf("Second message incorrect: %+v", history[1])
	}

	// Ensure history is a copy (modifying returned slice shouldn't affect internal)
	history[0].Content = "Modified"
	originalHistory := ic.GetHistory()
	if originalHistory[0].Content != "Hello" {
		t.Error("History should return a copy, not the original slice")
	}
}

func TestHandleCommand(t *testing.T) {
	tests := []struct {
		command     string
		expectError bool
		expectText  string
		verify      func(t *testing.T, ic *InteractiveChat)
	}{
		{"/help", false, "Available commands", nil},
		{"/clear", false, "Chat history cleared", func(t *testing.T, ic *InteractiveChat) {
			if len(ic.messages) != 0 {
				t.Errorf("Expected messages to be cleared, got %d", len(ic.messages))
			}
		}},
		{"/load test.json", false, "Load functionality not implemented", nil},
		{"/model use new-model", false, "Now using model", func(t *testing.T, ic *InteractiveChat) {
			if ic.model != "new-model" {
				t.Errorf("Expected model to switch to 'new-model', got '%s'", ic.model)
			}
			if len(ic.messages) != 0 {
				t.Errorf("Expected message history to reset after model switch, got %d", len(ic.messages))
			}
		}},
		{"/unknown", true, "", nil},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.command, func(t *testing.T) {
			var outputBuf strings.Builder

			ic := &InteractiveChat{
				writer:   &outputBuf,
				messages: []client.ChatMessage{{Role: "user", Content: "test"}},
				model:    "base-model",
			}

			err := ic.handleCommand(tc.command)

			if tc.expectError && err == nil {
				t.Fatalf("Expected error for command '%s', got none", tc.command)
			}

			if !tc.expectError && err != nil {
				t.Fatalf("Expected no error for command '%s', got: %v", tc.command, err)
			}

			if tc.expectText != "" {
				output := outputBuf.String()
				if !strings.Contains(output, tc.expectText) {
					t.Fatalf("Expected output to contain '%s', got: %s", tc.expectText, output)
				}
			}

			if tc.verify != nil {
				tc.verify(t, ic)
			}
		})
	}
}

func TestClearHistory(t *testing.T) {
	var outputBuf strings.Builder

	ic := &InteractiveChat{
		writer: &outputBuf,
		messages: []client.ChatMessage{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
	}

	if len(ic.messages) != 2 {
		t.Errorf("Expected 2 messages before clear, got %d", len(ic.messages))
	}

	err := ic.clearHistory()
	if err != nil {
		t.Errorf("Expected no error from clearHistory, got: %v", err)
	}

	if len(ic.messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(ic.messages))
	}

	output := outputBuf.String()
	if !strings.Contains(output, "Chat history cleared") {
		t.Errorf("Expected clear confirmation message, got: %s", output)
	}
}

func TestSaveCommandSavesHistoryToFile(t *testing.T) {
	var outputBuf strings.Builder
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "history.json")

	ic := &InteractiveChat{
		writer: &outputBuf,
		messages: []client.ChatMessage{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
	}

	cmd := fmt.Sprintf("/save %s", filePath)

	if err := ic.handleCommand(cmd); err != nil {
		t.Fatalf("handleCommand failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected history file to be created: %v", err)
	}

	var saved []client.ChatMessage
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("failed to unmarshal saved history: %v", err)
	}

	if len(saved) != len(ic.messages) {
		t.Fatalf("expected %d messages, got %d", len(ic.messages), len(saved))
	}

	if !strings.Contains(outputBuf.String(), "Chat history saved to") {
		t.Errorf("expected confirmation message, got: %s", outputBuf.String())
	}
}

func TestSaveCommandSavesPreviousResponse(t *testing.T) {
	var outputBuf strings.Builder
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "response.txt")

	ic := &InteractiveChat{
		writer: &outputBuf,
		messages: []client.ChatMessage{
			{Role: "user", Content: "Hi"},
			{Role: "assistant", Content: "Response content"},
		},
	}

	cmd := fmt.Sprintf("/save --previous --output %s", filePath)

	if err := ic.handleCommand(cmd); err != nil {
		t.Fatalf("handleCommand failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected response file to be created: %v", err)
	}

	if string(data) != "Response content" {
		t.Errorf("expected response content to be saved, got: %s", string(data))
	}

	if !strings.Contains(outputBuf.String(), "Saved previous response") {
		t.Errorf("expected confirmation message, got: %s", outputBuf.String())
	}
}

func TestSaveCommandPreviousResponseNoAssistant(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "response.txt")

	var outputBuf strings.Builder

	ic := &InteractiveChat{
		writer: &outputBuf,
		messages: []client.ChatMessage{
			{Role: "user", Content: "Only user"},
		},
	}

	cmd := fmt.Sprintf("/save --previous --output %s", filePath)

	if err := ic.handleCommand(cmd); err != nil {
		t.Fatalf("expected no error when saving without assistant message, got: %v", err)
	}

	if !strings.Contains(outputBuf.String(), "No assistant response available to save yet.") {
		t.Errorf("expected friendly notice, got: %s", outputBuf.String())
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("expected no file to be created, got err=%v", err)
	}
}

func TestShowHelpTTYAwareness(t *testing.T) {
	var outputBuf strings.Builder

	ic := &InteractiveChat{
		writer: &outputBuf,
	}

	if err := ic.showHelp(); err != nil {
		t.Fatalf("showHelp failed in non-TTY mode: %v", err)
	}

	if strings.Contains(outputBuf.String(), "\u001b[") {
		t.Error("Expected help output without ANSI codes in non-TTY mode")
	}

	outputBuf.Reset()

	ic.isTTY = true
	if err := ic.showHelp(); err != nil {
		t.Fatalf("showHelp failed in TTY mode: %v", err)
	}

	if !strings.Contains(outputBuf.String(), "\u001b[1;36m") {
		t.Error("Expected colored help output in TTY mode")
	}
}
