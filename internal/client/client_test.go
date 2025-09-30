package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	opts := Options{
		BaseURL: "http://localhost:11434",
		Token:   "test-token",
	}

	client := New(opts)

	if client.baseURL != opts.BaseURL {
		t.Errorf("Expected baseURL %s, got %s", opts.BaseURL, client.baseURL)
	}

	if client.token != opts.Token {
		t.Errorf("Expected token %s, got %s", opts.Token, client.token)
	}

	if client.retries != 3 {
		t.Errorf("Expected default retries 3, got %d", client.retries)
	}
}

func TestListModels(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("Expected path /api/tags, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		response := ListModelsResponse{
			Models: []Model{
				{
					Name: "llama2",
					Size: 4096000000,
				},
				{
					Name: "codellama",
					Size: 7000000000,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	resp, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(resp.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(resp.Models))
	}

	if resp.Models[0].Name != "llama2" {
		t.Errorf("Expected first model name 'llama2', got '%s'", resp.Models[0].Name)
	}

	if resp.Models[1].Name != "codellama" {
		t.Errorf("Expected second model name 'codellama', got '%s'", resp.Models[1].Name)
	}
}

func TestShowModel(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			t.Errorf("Expected path /api/show, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Decode request body
		var req ShowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Name != "llama2" {
			t.Errorf("Expected model name 'llama2', got '%s'", req.Name)
		}

		response := ShowResponse{
			License:   "MIT",
			Parameters: "temperature: 0.7",
			Details: ModelDetails{
				Format:        "gguf",
				ParameterSize: "7B",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := ShowRequest{Name: "llama2"}
	resp, err := client.ShowModel(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.License != "MIT" {
		t.Errorf("Expected license 'MIT', got '%s'", resp.License)
	}

	if resp.Details.ParameterSize != "7B" {
		t.Errorf("Expected parameter size '7B', got '%s'", resp.Details.ParameterSize)
	}
}

func TestGenerate(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path /api/generate, got %s", r.URL.Path)
		}

		// Decode request body
		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got '%s'", req.Model)
		}

		if req.Prompt != "Hello" {
			t.Errorf("Expected prompt 'Hello', got '%s'", req.Prompt)
		}

		response := GenerateResponse{
			Model:     "llama2",
			Response:  "Hello! How can I help you?",
			Done:      true,
			CreatedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := GenerateRequest{
		Model:  "llama2",
		Prompt: "Hello",
	}
	resp, err := client.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.Model != "llama2" {
		t.Errorf("Expected model 'llama2', got '%s'", resp.Model)
	}

	if resp.Response != "Hello! How can I help you?" {
		t.Errorf("Expected response 'Hello! How can I help you?', got '%s'", resp.Response)
	}

	if !resp.Done {
		t.Error("Expected Done to be true")
	}
}

func TestGenerateStream(t *testing.T) {
	// Mock server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		responses := []GenerateResponse{
			{Model: "llama2", Response: "Hello", Done: false},
			{Model: "llama2", Response: " there", Done: false},
			{Model: "llama2", Response: "!", Done: true},
		}

		for _, resp := range responses {
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		}
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := GenerateRequest{
		Model:  "llama2",
		Prompt: "Hello",
	}

	respCh, err := client.GenerateStream(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var responses []GenerateResponse
	for resp := range respCh {
		responses = append(responses, resp)
	}

	if len(responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(responses))
	}

	// Check that we received all chunks
	fullResponse := ""
	for _, resp := range responses {
		fullResponse += resp.Response
	}

	if fullResponse != "Hello there!" {
		t.Errorf("Expected full response 'Hello there!', got '%s'", fullResponse)
	}

	// Check that last response is marked as done
	if !responses[len(responses)-1].Done {
		t.Error("Last response should be marked as done")
	}
}

func TestClientWithToken(t *testing.T) {
	// Mock server that checks for Authorization header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", auth)
		}

		response := ListModelsResponse{Models: []Model{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		Token:   "test-token",
	})

	ctx := context.Background()
	_, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestClientErrorHandling(t *testing.T) {
	// Mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(Options{
		BaseURL: server.URL,
		Retries: 1, // Reduce retries for faster test
	})

	ctx := context.Background()
	_, err := client.ListModels(ctx)
	if err == nil {
		t.Error("Expected error for server error response")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to contain status code 500, got: %v", err)
	}
}