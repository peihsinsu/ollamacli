package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ollamacli/internal/client"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Options struct {
	Format Format
	Quiet  bool
	Writer io.Writer
}

type Formatter interface {
	FormatModels(models []client.Model) error
	FormatModelInfo(info *client.ShowResponse) error
	FormatChatResponse(resp *client.ChatResponse) error
	FormatGenerateResponse(resp *client.GenerateResponse) error
	FormatPullProgress(resp *client.PullResponse) error
	FormatEmbeddings(resp *client.EmbedResponse) error
	FormatCreateProgress(resp *client.CreateResponse) error
	FormatError(err error) error
	FormatGeneric(data interface{}) error
}

type formatter struct {
	opts Options
}

func New(opts Options) Formatter {
	return &formatter{opts: opts}
}

func (f *formatter) FormatModels(models []client.Model) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(map[string]interface{}{"models": models})
	default:
		return f.formatModelsText(models)
	}
}

func (f *formatter) formatModelsText(models []client.Model) error {
	if f.opts.Quiet {
		for _, model := range models {
			if _, err := fmt.Fprintln(f.opts.Writer, model.Name); err != nil {
				return err
			}
		}
		return nil
	}

	if len(models) == 0 {
		_, err := fmt.Fprintln(f.opts.Writer, "No models found")
		return err
	}

	_, err := fmt.Fprintf(f.opts.Writer, "%-30s %-15s %-10s %s\n", "NAME", "SIZE", "MODIFIED", "DIGEST")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f.opts.Writer, strings.Repeat("-", 80))
	if err != nil {
		return err
	}

	for _, model := range models {
		size := f.formatSize(model.Size)
		modified := model.ModifiedAt.Format("2006-01-02")
		digest := f.truncateString(model.Digest, 12)

		_, err = fmt.Fprintf(f.opts.Writer, "%-30s %-15s %-10s %s\n",
			model.Name, size, modified, digest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatModelInfo(info *client.ShowResponse) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(info)
	default:
		return f.formatModelInfoText(info)
	}
}

func (f *formatter) formatModelInfoText(info *client.ShowResponse) error {
	if f.opts.Quiet {
		_, err := fmt.Fprintln(f.opts.Writer, info.Details.ParameterSize)
		return err
	}

	sections := []struct {
		title   string
		content string
	}{
		{"License", info.License},
		{"Parameters", info.Parameters},
		{"Template", info.Template},
		{"Parameter Size", info.Details.ParameterSize},
		{"Format", info.Details.Format},
		{"Family", info.Details.Family},
		{"Quantization", info.Details.QuantizationLevel},
	}

	for _, section := range sections {
		if section.content != "" {
			_, err := fmt.Fprintf(f.opts.Writer, "%s:\n%s\n\n", section.title, section.content)
			if err != nil {
				return err
			}
		}
	}

	if info.Modelfile != "" {
		_, err := fmt.Fprintf(f.opts.Writer, "Modelfile:\n%s\n", info.Modelfile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *formatter) FormatChatResponse(resp *client.ChatResponse) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(resp)
	default:
		return f.formatChatResponseText(resp)
	}
}

func (f *formatter) formatChatResponseText(resp *client.ChatResponse) error {
	if f.opts.Quiet {
		_, err := fmt.Fprint(f.opts.Writer, resp.Message.Content)
		return err
	}

	_, err := fmt.Fprint(f.opts.Writer, resp.Message.Content)
	return err
}

func (f *formatter) FormatGenerateResponse(resp *client.GenerateResponse) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(resp)
	default:
		return f.formatGenerateResponseText(resp)
	}
}

func (f *formatter) formatGenerateResponseText(resp *client.GenerateResponse) error {
	if f.opts.Quiet {
		_, err := fmt.Fprint(f.opts.Writer, resp.Response)
		return err
	}

	_, err := fmt.Fprint(f.opts.Writer, resp.Response)
	return err
}

func (f *formatter) FormatPullProgress(resp *client.PullResponse) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(resp)
	default:
		return f.formatPullProgressText(resp)
	}
}

func (f *formatter) formatPullProgressText(resp *client.PullResponse) error {
	if f.opts.Quiet && !strings.HasPrefix(resp.Status, "error:") {
		return nil
	}

	status := resp.Status
	if strings.HasPrefix(status, "error:") {
		_, err := fmt.Fprintf(f.opts.Writer, "Error: %s\n", strings.TrimPrefix(status, "error: "))
		return err
	}

	if resp.Total > 0 && resp.Completed >= 0 {
		progress := float64(resp.Completed) / float64(resp.Total) * 100
		_, err := fmt.Fprintf(f.opts.Writer, "\r%s %.1f%% (%s/%s)",
			status, progress,
			f.formatSize(resp.Completed),
			f.formatSize(resp.Total))
		return err
	}

	_, err := fmt.Fprintf(f.opts.Writer, "%s\n", status)
	return err
}

func (f *formatter) FormatEmbeddings(resp *client.EmbedResponse) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(resp)
	default:
		return f.formatEmbeddingsText(resp)
	}
}

func (f *formatter) formatEmbeddingsText(resp *client.EmbedResponse) error {
	if f.opts.Quiet {
		for _, emb := range resp.Embeddings {
			_, err := fmt.Fprintf(f.opts.Writer, "%v\n", emb)
			if err != nil {
				return err
			}
		}
		return nil
	}

	_, err := fmt.Fprintf(f.opts.Writer, "Model: %s\n", resp.Model)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(f.opts.Writer, "Generated %d embedding(s)\n", len(resp.Embeddings))
	if err != nil {
		return err
	}

	for i, emb := range resp.Embeddings {
		_, err = fmt.Fprintf(f.opts.Writer, "Embedding %d: [%d dimensions]\n", i+1, len(emb))
		if err != nil {
			return err
		}
		if !f.opts.Quiet && len(emb) > 0 {
			preview := emb
			if len(preview) > 5 {
				preview = emb[:5]
			}
			_, err = fmt.Fprintf(f.opts.Writer, "  Preview: %v...\n", preview)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *formatter) FormatCreateProgress(resp *client.CreateResponse) error {
	switch f.opts.Format {
	case FormatJSON:
		return f.writeJSON(resp)
	default:
		return f.formatCreateProgressText(resp)
	}
}

func (f *formatter) formatCreateProgressText(resp *client.CreateResponse) error {
	if f.opts.Quiet && !strings.HasPrefix(resp.Status, "error:") {
		return nil
	}

	status := resp.Status
	if strings.HasPrefix(status, "error:") {
		_, err := fmt.Fprintf(f.opts.Writer, "Error: %s\n", strings.TrimPrefix(status, "error: "))
		return err
	}

	_, err := fmt.Fprintf(f.opts.Writer, "%s\n", status)
	return err
}

func (f *formatter) FormatError(err error) error {
	if f.opts.Format == FormatJSON {
		return f.writeJSON(map[string]string{"error": err.Error()})
	}

	_, writeErr := fmt.Fprintf(f.opts.Writer, "Error: %v\n", err)
	return writeErr
}

func (f *formatter) writeJSON(data interface{}) error {
	encoder := json.NewEncoder(f.opts.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// FormatGeneric formats any generic data structure
func (f *formatter) FormatGeneric(data interface{}) error {
	return f.writeJSON(data)
}

func (f *formatter) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (f *formatter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Streaming formatter for interactive use
type StreamFormatter struct {
	formatter
	buffer strings.Builder
}

func NewStream(opts Options) *StreamFormatter {
	return &StreamFormatter{
		formatter: formatter{opts: opts},
	}
}

func (sf *StreamFormatter) WriteChunk(chunk string) error {
	sf.buffer.WriteString(chunk)
	_, err := fmt.Fprint(sf.opts.Writer, chunk)
	return err
}

func (sf *StreamFormatter) GetBuffer() string {
	return sf.buffer.String()
}

func (sf *StreamFormatter) ClearBuffer() {
	sf.buffer.Reset()
}

func (sf *StreamFormatter) Flush() error {
	if sf.buffer.Len() > 0 && !sf.opts.Quiet {
		_, err := fmt.Fprintln(sf.opts.Writer)
		return err
	}
	return nil
}