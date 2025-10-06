package chat

import (
	"bufio"
	"context"
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
	"ollamacli/internal/rag"
)

// RAGInteractiveChat provides an interactive chat session with RAG support
type RAGInteractiveChat struct {
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
	retriever *rag.Retriever
	topK      int
}

// RAGOptions contains configuration for RAG interactive chat
type RAGOptions struct {
	Client    *client.Client
	Formatter output.Formatter
	Logger    log.Logger
	Model     string
	Writer    io.Writer
	Reader    io.Reader
	Prompt    string
	Retriever *rag.Retriever
	TopK      int
}

// NewRAGInteractiveChat creates a new RAG interactive chat session
func NewRAGInteractiveChat(opts RAGOptions) *RAGInteractiveChat {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.Reader == nil {
		opts.Reader = os.Stdin
	}
	if opts.Prompt == "" {
		opts.Prompt = DefaultPrompt
	}
	if opts.TopK == 0 {
		opts.TopK = 3
	}

	// Check if stdin is a TTY
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	opts.Logger.Debug("TTY detection: isTTY=%v, stdin_fd=%d", isTTY, os.Stdin.Fd())

	var line *liner.State
	var reader *bufio.Reader

	if isTTY {
		opts.Logger.Debug("Initializing liner for TTY mode")
		line = liner.NewLiner()
		line.SetCtrlCAborts(true)
	} else {
		opts.Logger.Debug("Falling back to bufio.Reader for non-TTY mode")
		reader = bufio.NewReader(opts.Reader)
	}

	return &RAGInteractiveChat{
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
		retriever: opts.Retriever,
		topK:      opts.TopK,
	}
}

// Start begins the interactive RAG chat session
func (ic *RAGInteractiveChat) Start(ctx context.Context) error {
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle signals
	go func() {
		<-sigChan
		ic.logger.Debug("Received interrupt signal, shutting down...")
		cancel()
		if ic.line != nil {
			ic.line.Close()
		}
		os.Exit(0)
	}()

	// Close liner on exit if TTY
	if ic.line != nil {
		defer ic.line.Close()
	}

	// Print welcome message
	fmt.Fprintf(ic.writer, "RAG Interactive Chat - Model: %s\n", ic.model)
	fmt.Fprintf(ic.writer, "Type %s for help, %s to exit\n\n", HelpCommand, ExitCommand)

	for {
		// Read user input
		var userInput string
		var err error

		if ic.isTTY {
			userInput, err = ic.line.Prompt(ic.prompt)
		} else {
			fmt.Fprint(ic.writer, ic.prompt)
			userInput, err = ic.reader.ReadString('\n')
			userInput = strings.TrimSpace(userInput)
		}

		if err != nil {
			if err == io.EOF || err == liner.ErrPromptAborted {
				fmt.Fprintln(ic.writer, "\nGoodbye!")
				return nil
			}
			return fmt.Errorf("failed to read input: %w", err)
		}

		userInput = strings.TrimSpace(userInput)

		// Skip empty input
		if userInput == "" {
			continue
		}

		// Add to history if TTY
		if ic.line != nil {
			ic.line.AppendHistory(userInput)
		}

		// Handle commands
		if strings.HasPrefix(userInput, "/") {
			if err := ic.handleCommand(ctx, userInput); err != nil {
				if err == io.EOF {
					return nil
				}
				fmt.Fprintf(ic.writer, "Error: %v\n", err)
			}
			continue
		}

		// Retrieve relevant context from knowledge base
		ic.logger.Debug("Retrieving relevant context for: %s", userInput)
		context, err := ic.retriever.RetrieveContext(ctx, userInput, ic.topK)
		if err != nil {
			ic.logger.Warn("Failed to retrieve context: %v", err)
			fmt.Fprintf(ic.writer, "Warning: Could not retrieve context from knowledge base\n")
		}

		// Construct augmented message
		augmentedInput := userInput
		if context != "" {
			augmentedInput = fmt.Sprintf("%s\n\nUser question: %s", context, userInput)
			ic.logger.Debug("Added context to query (%d chars)", len(context))
		} else {
			ic.logger.Debug("No relevant context found")
		}

		// Add system message if this is the first message
		if len(ic.messages) == 0 {
			ic.messages = append(ic.messages, client.ChatMessage{
				Role:    "system",
				Content: "You are a helpful assistant. Use the provided context to answer questions accurately. If the context doesn't contain relevant information, say so.",
			})
		}

		// Add user message with RAG context
		ic.messages = append(ic.messages, client.ChatMessage{
			Role:    "user",
			Content: augmentedInput,
		})

		// Create chat request
		req := client.ChatRequest{
			Model:    ic.model,
			Messages: ic.messages,
			Stream:   true,
		}

		// Send request and stream response
		respCh, err := ic.client.ChatStream(ctx, req)
		if err != nil {
			fmt.Fprintf(ic.writer, "Error: %v\n", err)
			// Remove the failed user message
			ic.messages = ic.messages[:len(ic.messages)-1]
			continue
		}

		// Collect full response
		var fullResponse strings.Builder
		for resp := range respCh {
			if err := ic.formatter.FormatChatResponse(&resp); err != nil {
				ic.logger.Warn("Failed to format response: %v", err)
			}
			if resp.Message.Content != "" {
				fullResponse.WriteString(resp.Message.Content)
			}
		}

		fmt.Fprintln(ic.writer) // New line after response

		// Add assistant response to conversation history
		ic.messages = append(ic.messages, client.ChatMessage{
			Role:    "assistant",
			Content: fullResponse.String(),
		})
	}
}

// handleCommand processes slash commands
func (ic *RAGInteractiveChat) handleCommand(ctx context.Context, cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case ExitCommand:
		fmt.Fprintln(ic.writer, "Goodbye!")
		return io.EOF

	case HelpCommand:
		ic.printHelp()
		return nil

	case ClearCommand:
		ic.messages = make([]client.ChatMessage, 0)
		fmt.Fprintln(ic.writer, "Conversation history cleared")
		return nil

	case StatusCommand:
		fmt.Fprintf(ic.writer, "Model: %s\n", ic.model)
		fmt.Fprintf(ic.writer, "Messages in context: %d\n", len(ic.messages))
		fmt.Fprintf(ic.writer, "RAG Top-K: %d\n", ic.topK)
		return nil

	default:
		fmt.Fprintf(ic.writer, "Unknown command: %s (type %s for help)\n", parts[0], HelpCommand)
		return nil
	}
}

// printHelp displays available commands
func (ic *RAGInteractiveChat) printHelp() {
	help := `
Available Commands:
  /help     - Show this help message
  /exit     - Exit the chat session
  /clear    - Clear conversation history
  /status   - Show current session status

RAG Features:
  - Each query automatically retrieves relevant context from the knowledge base
  - Conversation history is maintained across queries
  - Use 'ollamacli rag-import' to add documents to the knowledge base
`
	fmt.Fprintln(ic.writer, help)
}
