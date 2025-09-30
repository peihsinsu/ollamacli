package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/peterh/liner"
	"golang.org/x/term"
	"ollamacli/internal/client"
	"ollamacli/internal/log"
	"ollamacli/internal/output"
)

const (
	DefaultPrompt    = "> "
	ExitCommand      = "/exit"
	HelpCommand      = "/help"
	ClearCommand     = "/clear"
	SaveCommand      = "/save"
	LoadCommand      = "/load"
	ModelListCommand = "/model list"
	ModelPullCommand = "/model pull"
	ModelShowCommand = "/model show"
	ModelUseCommand  = "/model use"
	StatusCommand    = "/status"
)

type InteractiveChat struct {
	client    *client.Client
	formatter output.Formatter
	logger    log.Logger
	model     string
	messages  []client.ChatMessage
	line      *liner.State
	reader    *bufio.Reader
	writer    io.Writer
	prompt    string
	isTTY     bool
}

type Options struct {
	Client    *client.Client
	Formatter output.Formatter
	Logger    log.Logger
	Model     string
	Writer    io.Writer
	Reader    io.Reader
	Prompt    string
}

func NewInteractiveChat(opts Options) *InteractiveChat {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.Reader == nil {
		opts.Reader = os.Stdin
	}
	if opts.Prompt == "" {
		opts.Prompt = DefaultPrompt
	}

	// Check if stdin is a TTY
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	// Debug output
	opts.Logger.Debug("TTY detection: isTTY=%v, stdin_fd=%d", isTTY, os.Stdin.Fd())

	var line *liner.State
	var reader *bufio.Reader

	if isTTY {
		// Use liner for TTY with readline support
		opts.Logger.Debug("Initializing liner for TTY mode")
		line = liner.NewLiner()
		line.SetCtrlCAborts(true)
	} else {
		// Fallback to bufio.Reader for non-TTY
		opts.Logger.Debug("Falling back to bufio.Reader for non-TTY mode")
		reader = bufio.NewReader(opts.Reader)
	}

	return &InteractiveChat{
		client:    opts.Client,
		formatter: opts.Formatter,
		logger:    opts.Logger,
		model:     opts.Model,
		messages:  make([]client.ChatMessage, 0),
		line:      line,
		reader:    reader,
		writer:    opts.Writer,
		prompt:    opts.Prompt,
		isTTY:     isTTY,
	}
}

func (ic *InteractiveChat) Start(ctx context.Context) error {
	ic.logger.Info("Starting interactive chat with model: %s", ic.model)

	// Ensure liner is closed on exit (if TTY)
	if ic.isTTY && ic.line != nil {
		defer ic.line.Close()
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintf(ic.writer, "\nGoodbye! Session ended.\n")
		if ic.isTTY && ic.line != nil {
			ic.line.Close()
		}
		cancel()
	}()

	// Welcome message with colors (only if TTY)
	if ic.isTTY {
		fmt.Fprintf(ic.writer, "\033[1;36mInteractive mode with %s\033[0m (type \033[1;33m/help\033[0m for commands, Ctrl+C to exit)\n\n", ic.model)
	} else {
		fmt.Fprintf(ic.writer, "Interactive mode with %s (type /help for commands, Ctrl+C to exit)\n\n", ic.model)
	}

	// Main chat loop
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var input string
		var err error

		// Read user input (different method for TTY vs non-TTY)
		if ic.isTTY {
			// TTY mode: use liner without ANSI codes since liner rejects control chars
			ic.logger.Debug("Using liner.Prompt() for input")
			input, err = ic.line.Prompt(ic.prompt)
			if err != nil {
				ic.logger.Debug("liner.Prompt() error: %v", err)
				if err == liner.ErrPromptAborted || err == io.EOF {
					fmt.Fprintf(ic.writer, "\nGoodbye! Session ended.\n")
					return nil
				}
				return fmt.Errorf("failed to read input: %w", err)
			}
			ic.logger.Debug("liner.Prompt() success, input length: %d", len(input))
		} else {
			// Non-TTY mode: use bufio.Reader
			fmt.Fprint(ic.writer, ic.prompt)
			input, err = ic.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Fprintf(ic.writer, "\nGoodbye! Session ended.\n")
					return nil
				}
				return fmt.Errorf("failed to read input: %w", err)
			}
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Add to history (only for TTY with liner)
		if ic.isTTY {
			ic.line.AppendHistory(input)
		}

		// Handle "exit" or "quit" without slash
		if input == "exit" || input == "quit" {
			fmt.Fprintf(ic.writer, "Goodbye! Session ended.\n")
			return nil
		}

		// Handle special commands
		if strings.HasPrefix(input, "/") {
			if err := ic.handleCommand(input); err != nil {
				fmt.Fprintf(ic.writer, "\033[1;31mError:\033[0m %v\n", err)
			}
			continue
		}

		// Send message to model
		if err := ic.sendMessage(ctx, input); err != nil {
			fmt.Fprintf(ic.writer, "\033[1;31mError:\033[0m %v\n", err)
			continue
		}
	}
}

func (ic *InteractiveChat) handleCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	// Handle multi-word commands like "/model list"
	fullCmd := command
	if strings.HasPrefix(command, "/model ") && len(parts) >= 2 {
		fullCmd = strings.Join(parts[0:2], " ")
	}

	cmd := parts[0]
	args := parts[1:]

	switch {
	case cmd == HelpCommand:
		return ic.showHelp()
	case cmd == ClearCommand:
		return ic.clearHistory()
	case fullCmd == ModelListCommand:
		return ic.modelList()
	case fullCmd == ModelPullCommand:
		if len(parts) < 3 {
			return fmt.Errorf("usage: /model pull <model_name>")
		}
		return ic.modelPull(parts[2])
	case fullCmd == ModelUseCommand:
		if len(parts) < 3 {
			return fmt.Errorf("usage: /model use <model_name>")
		}
		return ic.modelUse(parts[2])
	case fullCmd == ModelShowCommand:
		if len(parts) < 3 {
			return fmt.Errorf("usage: /model show <model_name>")
		}
		return ic.modelShow(parts[2])
	case cmd == StatusCommand:
		return ic.showStatus()
	case cmd == SaveCommand:
		return ic.saveCommand(args)
	case cmd == LoadCommand:
		filename := "chat_history.json"
		if len(args) > 0 {
			filename = args[0]
		}
		return ic.loadHistory(filename)
	case cmd == ExitCommand:
		fmt.Fprintf(ic.writer, "Goodbye! Session ended.\n")
		os.Exit(0)
		return nil
	default:
		return fmt.Errorf("unknown command: %s (type /help for available commands)", cmd)
	}
}

func (ic *InteractiveChat) showHelp() error {
	headerColor := ""
	commandColor := ""
	tipColor := ""
	resetColor := ""

	if ic.isTTY {
		headerColor = "\033[1;36m"
		commandColor = "\033[1;33m"
		tipColor = "\033[1;32m"
		resetColor = "\033[0m"
	}

	help := fmt.Sprintf(`%sAvailable commands:%s
  %s/help%s                    - Show this help message
  %s/clear%s                   - Clear chat history
  %s/model list%s              - List available models
  %s/model pull%s <name>       - Pull a model from registry
  %s/model use%s <name>        - Switch the active model
  %s/model show%s <name>       - Show model information
  %s/status%s                  - Show current session status
  %s/save%s [filename]         - Save chat history (default: chat_history.json)
  %s/save%s --previous --output <path> - Save the last response to file
  %s/load%s [filename]         - Load chat history (default: chat_history.json)
  %s/exit%s                    - Exit the chat

%sTips:%s
  - Use %sUp/Down arrows%s to navigate command history
  - Press %sCtrl+C%s to exit gracefully
`,
		headerColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		commandColor, resetColor,
		headerColor, resetColor,
		tipColor, resetColor,
		tipColor, resetColor)

	_, err := fmt.Fprint(ic.writer, help)
	return err
}

func (ic *InteractiveChat) modelList() error {
	ic.logger.Debug("Listing models from interactive mode")

	resp, err := ic.client.ListModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(resp.Models) == 0 {
		fmt.Fprintln(ic.writer, "\033[1;33mNo models found\033[0m")
		return nil
	}

	fmt.Fprintln(ic.writer, "\033[1;36mAvailable models:\033[0m")
	for _, model := range resp.Models {
		fmt.Fprintf(ic.writer, "  \033[1;32m•\033[0m %s\n", model.Name)
	}
	fmt.Fprintln(ic.writer)

	return nil
}

func (ic *InteractiveChat) modelPull(modelName string) error {
	ic.logger.Debug("Pulling model: %s", modelName)

	fmt.Fprintf(ic.writer, "Pulling model: %s\n", modelName)

	req := client.PullRequest{
		Name:   modelName,
		Stream: true,
	}

	respCh, err := ic.client.PullModel(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to start pull: %w", err)
	}

	for resp := range respCh {
		if strings.HasPrefix(resp.Status, "error:") {
			return fmt.Errorf("pull failed: %s", resp.Status)
		}

		if resp.Total > 0 && resp.Completed >= 0 {
			progress := float64(resp.Completed) / float64(resp.Total) * 100
			fmt.Fprintf(ic.writer, "\r%s %.1f%%", resp.Status, progress)
		} else {
			fmt.Fprintf(ic.writer, "\r%s", resp.Status)
		}
	}

	fmt.Fprintf(ic.writer, "\nModel pulled successfully!\n\n")
	return nil
}

func (ic *InteractiveChat) modelUse(modelName string) error {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return fmt.Errorf("usage: /model use <model_name>")
	}

	if ic.logger != nil {
		ic.logger.Info("Switching interactive model to: %s", modelName)
	}

	highlight := ""
	reset := ""
	if ic.isTTY {
		highlight = "\033[1;36m"
		reset = "\033[0m"
	}

	if ic.model == modelName {
		fmt.Fprintf(ic.writer, "%sAlready using model:%s %s\n\n", highlight, reset, modelName)
		return nil
	}

	ic.model = modelName
	ic.messages = make([]client.ChatMessage, 0)

	fmt.Fprintf(ic.writer, "%sNow using model:%s %s\n", highlight, reset, modelName)
	fmt.Fprintf(ic.writer, "Chat history cleared for new model.\n\n")
	return nil
}

func (ic *InteractiveChat) modelShow(modelName string) error {
	ic.logger.Debug("Showing model info: %s", modelName)

	req := client.ShowRequest{Name: modelName}
	resp, err := ic.client.ShowModel(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to show model: %w", err)
	}

	fmt.Fprintf(ic.writer, "\033[1;36mModel:\033[0m %s\n", modelName)
	if resp.Details.ParameterSize != "" {
		fmt.Fprintf(ic.writer, "  \033[1;33mParameter Size:\033[0m %s\n", resp.Details.ParameterSize)
	}
	if resp.Details.Format != "" {
		fmt.Fprintf(ic.writer, "  \033[1;33mFormat:\033[0m %s\n", resp.Details.Format)
	}
	if resp.Details.Family != "" {
		fmt.Fprintf(ic.writer, "  \033[1;33mFamily:\033[0m %s\n", resp.Details.Family)
	}
	if resp.Details.QuantizationLevel != "" {
		fmt.Fprintf(ic.writer, "  \033[1;33mQuantization:\033[0m %s\n", resp.Details.QuantizationLevel)
	}
	fmt.Fprintln(ic.writer)

	return nil
}

func (ic *InteractiveChat) showStatus() error {
	fmt.Fprintln(ic.writer, "\033[1;36mSession Status:\033[0m")
	fmt.Fprintf(ic.writer, "  \033[1;33mModel:\033[0m %s\n", ic.model)
	fmt.Fprintf(ic.writer, "  \033[1;33mTotal messages:\033[0m %d\n", len(ic.messages))

	// Count user and assistant messages separately
	userMsgs := 0
	assistantMsgs := 0
	for _, msg := range ic.messages {
		if msg.Role == "user" {
			userMsgs++
		} else if msg.Role == "assistant" {
			assistantMsgs++
		}
	}

	fmt.Fprintf(ic.writer, "  \033[1;33mUser messages:\033[0m %d\n", userMsgs)
	fmt.Fprintf(ic.writer, "  \033[1;33mAssistant messages:\033[0m %d\n", assistantMsgs)
	fmt.Fprintln(ic.writer)
	return nil
}

func (ic *InteractiveChat) clearHistory() error {
	ic.messages = make([]client.ChatMessage, 0)
	_, err := fmt.Fprintf(ic.writer, "Chat history cleared.\n")
	return err
}

func (ic *InteractiveChat) saveCommand(args []string) error {
	options := struct {
		filename string
		output   string
		previous bool
	}{
		filename: "chat_history.json",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--previous":
			options.previous = true
		case "--output":
			i++
			if i >= len(args) {
				return fmt.Errorf("missing value for --output")
			}
			options.output = args[i]
		default:
			if strings.HasPrefix(arg, "--") {
				return fmt.Errorf("unknown flag for /save: %s", arg)
			}
			if options.filename != "chat_history.json" && options.output == "" {
				return fmt.Errorf("multiple filenames provided: %s", arg)
			}
			options.filename = arg
		}
	}

	if options.previous {
		if options.output == "" {
			return fmt.Errorf("--output <path> is required when using --previous")
		}
		content, err := ic.lastAssistantMessage()
		if err != nil {
			fmt.Fprintln(ic.writer, "No assistant response available to save yet.")
			return nil
		}
		if err := os.WriteFile(options.output, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to save previous response: %w", err)
		}
		_, err = fmt.Fprintf(ic.writer, "Saved previous response to %s\n", options.output)
		return err
	}

	target := options.output
	if target == "" {
		target = options.filename
	}

	data, err := json.MarshalIndent(ic.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode chat history: %w", err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("failed to save chat history: %w", err)
	}
	_, err = fmt.Fprintf(ic.writer, "Chat history saved to %s\n", target)
	return err
}

func (ic *InteractiveChat) lastAssistantMessage() (string, error) {
	for i := len(ic.messages) - 1; i >= 0; i-- {
		msg := ic.messages[i]
		if msg.Role == "assistant" {
			if msg.Content == "" {
				return "", fmt.Errorf("last assistant response is empty")
			}
			return msg.Content, nil
		}
	}
	return "", fmt.Errorf("no assistant response available to save")
}

func (ic *InteractiveChat) loadHistory(filename string) error {
	// TODO: Implement load functionality
	_, err := fmt.Fprintf(ic.writer, "Load functionality not implemented yet: %s\n", filename)
	return err
}

func (ic *InteractiveChat) sendMessage(ctx context.Context, message string) error {
	// Add user message to history
	userMsg := client.ChatMessage{
		Role:    "user",
		Content: message,
	}
	ic.messages = append(ic.messages, userMsg)

	// Prepare request
	req := client.ChatRequest{
		Model:    ic.model,
		Messages: ic.messages,
		Stream:   true,
	}

	// Send request and handle streaming response
	respCh, err := ic.client.ChatStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start chat stream: %w", err)
	}

	// Process streaming responses
	var assistantMsg client.ChatMessage
	var responseBuilder strings.Builder

	for resp := range respCh {
		if resp.Done {
			// Final response
			if resp.Message.Content != "" {
				responseBuilder.WriteString(resp.Message.Content)
			}
			break
		}

		// Stream response chunk
		chunk := resp.Message.Content
		responseBuilder.WriteString(chunk)
		fmt.Fprint(ic.writer, chunk)
	}

	fmt.Fprintf(ic.writer, "\n\n") // Two new lines after response for next prompt

	// Add assistant message to history
	assistantMsg = client.ChatMessage{
		Role:    "assistant",
		Content: responseBuilder.String(),
	}
	ic.messages = append(ic.messages, assistantMsg)

	return nil
}

func (ic *InteractiveChat) GetHistory() []client.ChatMessage {
	// Return a copy to prevent external modification
	history := make([]client.ChatMessage, len(ic.messages))
	copy(history, ic.messages)
	return history
}

func (ic *InteractiveChat) SetHistory(messages []client.ChatMessage) {
	ic.messages = make([]client.ChatMessage, len(messages))
	copy(ic.messages, messages)
}
