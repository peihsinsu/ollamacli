package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	retries    int
	retryDelay time.Duration
}

type Options struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	Retries    int
	RetryDelay time.Duration
}

func New(opts Options) *Client {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.Retries == 0 {
		opts.Retries = 3
	}
	if opts.RetryDelay == 0 {
		opts.RetryDelay = 5 * time.Second
	}

	return &Client{
		baseURL: opts.BaseURL,
		token:   opts.Token,
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		retries:    opts.Retries,
		retryDelay: opts.RetryDelay,
	}
}

type Model struct {
	Name         string            `json:"name"`
	ModifiedAt   time.Time         `json:"modified_at"`
	Size         int64             `json:"size"`
	Digest       string            `json:"digest"`
	Details      ModelDetails      `json:"details"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	SizeVRAM     int64             `json:"size_vram,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Families     []string          `json:"families,omitempty"`
	ParentModel  string            `json:"parent_model,omitempty"`
	Format       string            `json:"format,omitempty"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type ListModelsResponse struct {
	Models []Model `json:"models"`
}

type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Raw      bool                   `json:"raw,omitempty"`
}

type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []ChatMessage          `json:"messages"`
	Stream   bool                   `json:"stream,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Model              string      `json:"model"`
	CreatedAt          time.Time   `json:"created_at"`
	Message            ChatMessage `json:"message"`
	Done               bool        `json:"done"`
	TotalDuration      int64       `json:"total_duration,omitempty"`
	LoadDuration       int64       `json:"load_duration,omitempty"`
	PromptEvalCount    int         `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64       `json:"prompt_eval_duration,omitempty"`
	EvalCount          int         `json:"eval_count,omitempty"`
	EvalDuration       int64       `json:"eval_duration,omitempty"`
}

type PullRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type ShowRequest struct {
	Name string `json:"name"`
}

type ShowResponse struct {
	License    string       `json:"license,omitempty"`
	Modelfile  string       `json:"modelfile,omitempty"`
	Parameters string       `json:"parameters,omitempty"`
	Template   string       `json:"template,omitempty"`
	Details    ModelDetails `json:"details,omitempty"`
}

type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

type CreateRequest struct {
	Name      string `json:"name"`
	Modelfile string `json:"modelfile,omitempty"`
	Path      string `json:"path,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
	Quantize  string `json:"quantize,omitempty"`
}

type CreateResponse struct {
	Status string `json:"status"`
}

func (c *Client) ListModels(ctx context.Context) (*ListModelsResponse, error) {
	var result ListModelsResponse
	err := c.doRequest(ctx, "GET", "/api/tags", nil, &result)
	return &result, err
}

func (c *Client) PullModel(ctx context.Context, req PullRequest) (<-chan PullResponse, error) {
	req.Stream = true
	respCh := make(chan PullResponse)

	go func() {
		defer close(respCh)
		err := c.streamRequest(ctx, "POST", "/api/pull", req, func(data []byte) error {
			var resp PullResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}
			select {
			case respCh <- resp:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
		if err != nil {
			// Send error as status
			select {
			case respCh <- PullResponse{Status: fmt.Sprintf("error: %v", err)}:
			case <-ctx.Done():
			}
		}
	}()

	return respCh, nil
}

func (c *Client) ShowModel(ctx context.Context, req ShowRequest) (*ShowResponse, error) {
	var result ShowResponse
	err := c.doRequest(ctx, "POST", "/api/show", req, &result)
	return &result, err
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	var result GenerateResponse
	err := c.doRequest(ctx, "POST", "/api/generate", req, &result)
	return &result, err
}

func (c *Client) GenerateStream(ctx context.Context, req GenerateRequest) (<-chan GenerateResponse, error) {
	req.Stream = true
	respCh := make(chan GenerateResponse)

	go func() {
		defer close(respCh)
		err := c.streamRequest(ctx, "POST", "/api/generate", req, func(data []byte) error {
			var resp GenerateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}
			select {
			case respCh <- resp:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
		if err != nil && err != ctx.Err() {
			// Send error in response
			select {
			case respCh <- GenerateResponse{Response: fmt.Sprintf("Error: %v", err), Done: true}:
			case <-ctx.Done():
			}
		}
	}()

	return respCh, nil
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var result ChatResponse
	err := c.doRequest(ctx, "POST", "/api/chat", req, &result)
	return &result, err
}

func (c *Client) ChatStream(ctx context.Context, req ChatRequest) (<-chan ChatResponse, error) {
	req.Stream = true
	respCh := make(chan ChatResponse)

	go func() {
		defer close(respCh)
		err := c.streamRequest(ctx, "POST", "/api/chat", req, func(data []byte) error {
			var resp ChatResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}
			select {
			case respCh <- resp:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
		if err != nil && err != ctx.Err() {
			// Send error in message
			select {
			case respCh <- ChatResponse{
				Message: ChatMessage{Role: "assistant", Content: fmt.Sprintf("Error: %v", err)},
				Done:    true,
			}:
			case <-ctx.Done():
			}
		}
	}()

	return respCh, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	var resp *http.Response
	for attempt := 0; attempt <= c.retries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if attempt < c.retries {
			time.Sleep(c.retryDelay)
		}
	}

	if err != nil {
		return fmt.Errorf("request failed after %d retries: %w", c.retries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respData))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Embed(ctx context.Context, req EmbedRequest) (*EmbedResponse, error) {
	var result EmbedResponse
	err := c.doRequest(ctx, "POST", "/api/embed", req, &result)
	return &result, err
}

func (c *Client) CreateModel(ctx context.Context, req CreateRequest) (<-chan CreateResponse, error) {
	req.Stream = true
	respCh := make(chan CreateResponse)

	go func() {
		defer close(respCh)
		err := c.streamRequest(ctx, "POST", "/api/create", req, func(data []byte) error {
			var resp CreateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}
			select {
			case respCh <- resp:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
		if err != nil {
			select {
			case respCh <- CreateResponse{Status: fmt.Sprintf("error: %v", err)}:
			case <-ctx.Done():
			}
		}
	}()

	return respCh, nil
}

func (c *Client) streamRequest(ctx context.Context, method, path string, reqBody interface{}, handler func([]byte) error) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respData))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to decode stream response: %w", err)
		}

		if err := handler(raw); err != nil {
			return err
		}
	}
}