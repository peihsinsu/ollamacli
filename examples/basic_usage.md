# ollamacli Examples

This document provides examples of how to use the ollamacli tool.

## Prerequisites

Before using ollamacli, ensure that:
1. You have an Ollama server running (default: localhost:11434)
2. You have at least one model installed on your Ollama server

## Basic Commands

### List Available Models

```bash
# List all available models
ollamacli list

# List models in JSON format
ollamacli list --format json

# List models quietly (names only)
ollamacli list --quiet
```

### Pull a Model

```bash
# Pull a model from the registry
ollamacli pull llama2

# Pull a model with retry settings
ollamacli pull llama2 --retry 5 --retry-delay 10

# Pull a model allowing insecure connections
ollamacli pull llama2 --insecure
```

### Show Model Information

```bash
# Show detailed information about a model
ollamacli show llama2

# Show model info in JSON format
ollamacli show llama2 --format json

# Show model info quietly (minimal output)
ollamacli show llama2 --quiet
```

## Chat and Generation

### Single Response Chat

```bash
# Chat with a model using a prompt
ollamacli chat llama2 --prompt "Hello, how are you?"

# Chat with input from stdin
echo "What is the capital of France?" | ollamacli chat llama2

# Chat with JSON output
ollamacli chat llama2 --prompt "Explain JSON" --format json
```

### Interactive Chat Mode

```bash
# Start interactive chat session
ollamacli chat llama2 --interactive

# Interactive chat with specific host
ollamacli chat llama2 --interactive --host remote-server --port 11435
```

#### Interactive Mode Commands

Once in interactive mode, you can use these commands:
- `/help` - Show available commands
- `/clear` - Clear chat history
- `/save [filename]` - Save chat history (not yet implemented)
- `/load [filename]` - Load chat history (not yet implemented)
- `/exit` - Exit interactive mode
- `Ctrl+C` - Exit gracefully

#### Interactive Mode Example Session

```
$ ollamacli chat llama2 --interactive
Connected to llama2 model. Type your message or press Ctrl+C to exit.

> Hello! What's the weather like?
I don't have access to real-time weather data. You might want to check a weather app or website for current conditions in your area.

> What can you help me with?
I can help you with a variety of tasks including:
- Answering questions on many topics
- Writing and editing text
- Programming assistance
- Problem solving
- Creative writing
- And much more!

> /help
Available commands:
  /help               - Show this help message
  /clear              - Clear chat history
  /save [filename]    - Save chat history (default: chat_history.json)
  /load [filename]    - Load chat history (default: chat_history.json)
  /exit               - Exit the chat

> /exit
Goodbye! Session ended.
```

### Single Generation (Run Command)

```bash
# Run a single generation
ollamacli run llama2 --prompt "Write a haiku about programming"

# Run with input from file
ollamacli run llama2 --prompt "$(cat prompt.txt)"

# Run with stdin input
cat document.txt | ollamacli run llama2 --prompt "Summarize this:"
```

## Configuration

### Environment Variables

```bash
# Set server host and port
export OLLAMA_HOST=remote-server.example.com
export OLLAMA_PORT=11434

# Set authentication token
export OLLAMA_TOKEN=your-auth-token

# Set logging level
export OLLAMA_LOG_LEVEL=debug
export OLLAMA_VERBOSE=true

# Use the CLI with environment settings
ollamacli list
```

### Command-line Flags

```bash
# Override host and port
ollamacli list --host localhost --port 11434

# Enable verbose logging
ollamacli list --verbose

# Enable quiet mode
ollamacli list --quiet

# Use authentication token
ollamacli list --token your-auth-token
```

### Configuration File

Create a configuration file at `~/.ollamacli/config.yaml`:

```yaml
host: localhost
port: 11434
token: your-auth-token
log_level: info
verbose: false
quiet: false
```

## Advanced Usage

### Piping and Scripting

```bash
# Get list of model names for scripting
ollamacli list --quiet

# Use in a shell script
#!/bin/bash
MODELS=$(ollamacli list --quiet)
for model in $MODELS; do
    echo "Testing model: $model"
    ollamacli chat $model --prompt "Say hello" --quiet
done

# Process multiple prompts
cat prompts.txt | while read prompt; do
    ollamacli run llama2 --prompt "$prompt" >> responses.txt
done
```

### Error Handling

```bash
# Check exit codes
ollamacli list
if [ $? -eq 0 ]; then
    echo "Success"
else
    echo "Failed with exit code $?"
fi

# Common exit codes:
# 0: Success
# 1: General error
# 2: Usage error (invalid arguments)
# 3: Connection error (cannot reach Ollama server)
# 4: Authentication error
```

### JSON Output Processing

```bash
# Extract model names using jq
ollamacli list --format json | jq -r '.models[].name'

# Get model sizes
ollamacli list --format json | jq -r '.models[] | "\(.name): \(.size)"'

# Pretty print model information
ollamacli show llama2 --format json | jq .
```

## Troubleshooting

### Connection Issues

```bash
# Test connection with verbose output
ollamacli list --verbose

# Try with different host/port
ollamacli list --host 127.0.0.1 --port 11434

# Check if Ollama server is running
curl http://localhost:11434/api/tags
```

### Model Issues

```bash
# Check if model exists
ollamacli show llama2

# Pull model if missing
ollamacli pull llama2

# List all available models
ollamacli list
```

### Debug Mode

```bash
# Enable debug logging
export OLLAMA_LOG_LEVEL=debug
export OLLAMA_VERBOSE=true
ollamacli chat llama2 --prompt "test"

# Or use flags
ollamacli chat llama2 --prompt "test" --verbose
```

## Integration Examples

### With curl

```bash
# Compare ollamacli with direct API calls
# Using ollamacli:
ollamacli chat llama2 --prompt "Hello"

# Using curl directly:
curl http://localhost:11434/api/chat -d '{
  "model": "llama2",
  "messages": [{"role": "user", "content": "Hello"}],
  "stream": false
}'
```

### With Docker

```bash
# Run ollamacli in a Docker container (hypothetical)
docker run --rm -it \
  -e OLLAMA_HOST=host.docker.internal \
  ollamacli/ollamacli chat llama2 --prompt "Hello from Docker"
```

### With CI/CD

```bash
# Example GitHub Actions workflow step
- name: Test AI Model
  run: |
    ollamacli chat llama2 --prompt "Analyze this code" --quiet > analysis.txt
    if [ -s analysis.txt ]; then
      echo "AI analysis completed"
    else
      echo "AI analysis failed"
      exit 1
    fi
```

## Performance Tips

1. **Use `--quiet` mode for scripting** to reduce output overhead
2. **Set appropriate timeouts** for long-running operations
3. **Use non-interactive mode** for automated tasks
4. **Consider model size** when choosing models for frequent operations
5. **Reuse interactive sessions** for multiple queries to the same model