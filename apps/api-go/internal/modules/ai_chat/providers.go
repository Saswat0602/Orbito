package ai_chat

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strings"

	"github.com/orbito/api-go/internal/platform/safehttp"
	"github.com/orbito/api-go/internal/shared/apperror"
)

var validModelNameRe = regexp.MustCompile(`^[A-Za-z0-9._:/-]+$`)

// NormalizeMessages removes blank messages and merges consecutive same-role messages.
func NormalizeMessages(messages []ChatMessage) []ChatMessage {
	out := make([]ChatMessage, 0, len(messages))
	for _, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		n := len(out)
		if n > 0 && out[n-1].Role == msg.Role {
			out[n-1].Content = out[n-1].Content + "\n\n" + msg.Content
		} else {
			out = append(out, ChatMessage{Role: msg.Role, Content: msg.Content})
		}
	}
	return out
}

// DetectProvider identifies the LLM provider dialect from the configured base URL.
func DetectProvider(apiURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(apiURL))
	if err != nil {
		return "custom"
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return "ollama"
	}
	if addr, pErr := netip.ParseAddr(host); pErr == nil && safehttp.IsPrivateIP(addr) {
		return "ollama"
	}

	switch {
	case host == "openrouter.ai" || strings.HasSuffix(host, ".openrouter.ai"):
		return "openrouter"
	case host == "api.openai.com" || strings.HasSuffix(host, ".api.openai.com"):
		return "openai"
	case host == "api.anthropic.com" || strings.HasSuffix(host, ".api.anthropic.com"):
		return "anthropic"
	case host == "generativelanguage.googleapis.com" || strings.HasSuffix(host, ".generativelanguage.googleapis.com"):
		return "google"
	case host == "api.cohere.com" || strings.HasSuffix(host, ".api.cohere.com"):
		return "cohere"
	case host == "api.groq.com" || strings.HasSuffix(host, ".groq.com"):
		return "groq"
	case strings.Contains(host, "bedrock") || strings.Contains(host, "bedrock-runtime"):
		return "bedrock"
	default:
		return "custom"
	}
}

// ValidateModelName rejects characters that could inject URL path segments or query parameters.
func ValidateModelName(model string) error {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" || !validModelNameRe.MatchString(trimmed) {
		return apperror.BadRequest("Invalid AI model name")
	}
	return nil
}

// BuildProviderRequest formats endpoint URL, headers, and JSON body for the detected dialect.
func BuildProviderRequest(
	provider, apiURL, apiKey, model, appURL string,
	messages []ChatMessage,
	maxTokens int,
	temperature float64,
	samplingExtras bool,
) (string, map[string]string, map[string]any, error) {
	base := strings.TrimRight(apiURL, "/")
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + apiKey,
	}
	body := map[string]any{
		"model":       model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      false,
	}

	switch provider {
	case "openrouter":
		if appURL == "" {
			appURL = "http://localhost:3000"
		}
		headers["HTTP-Referer"] = appURL
		headers["X-Title"] = "Orbito AI Assistant"
		if samplingExtras {
			body["top_p"] = 0.9
			body["frequency_penalty"] = 0
			body["presence_penalty"] = 0
		}
		return base + "/chat/completions", headers, body, nil

	case "groq", "bedrock":
		if samplingExtras {
			body["top_p"] = 0.9
		}
		return base + "/chat/completions", headers, body, nil

	case "openai":
		delete(body, "max_tokens")
		body["max_completion_tokens"] = maxTokens
		if strings.HasPrefix(model, "gpt-5") {
			delete(body, "temperature")
		} else if samplingExtras {
			body["top_p"] = 0.9
			body["frequency_penalty"] = 0
			body["presence_penalty"] = 0
		}
		return base + "/chat/completions", headers, body, nil

	case "ollama":
		delete(headers, "Authorization")
		if samplingExtras {
			body["top_p"] = 0.9
		}
		return resolveOllamaEndpoint(base), headers, body, nil

	case "anthropic":
		delete(headers, "Authorization")
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
		return base + "/messages", headers, buildAnthropicBody(model, messages, maxTokens, temperature), nil

	case "google":
		if err := ValidateModelName(model); err != nil {
			return "", nil, nil, err
		}
		delete(headers, "Authorization")
		reqURL := fmt.Sprintf("%s/models/%s:generateContent?key=%s", base, url.PathEscape(strings.TrimSpace(model)), url.QueryEscape(apiKey))
		return reqURL, headers, buildGoogleBody(messages, maxTokens, temperature), nil

	default:
		return base + "/chat/completions", headers, body, nil
	}
}

func resolveOllamaEndpoint(base string) string {
	switch {
	case strings.Contains(base, "/v1"):
		if strings.HasSuffix(base, "/chat/completions") {
			return base
		}
		return base + "/chat/completions"
	case strings.Contains(base, "/api"):
		if strings.HasSuffix(base, "/chat") {
			return base
		}
		return base + "/chat"
	default:
		return base + "/v1/chat/completions"
	}
}

func buildAnthropicBody(model string, messages []ChatMessage, maxTokens int, temperature float64) map[string]any {
	var sys string
	filtered := make([]ChatMessage, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			sys = m.Content
			continue
		}
		filtered = append(filtered, m)
	}
	res := map[string]any{
		"model":       model,
		"messages":    filtered,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}
	if sys != "" {
		res["system"] = sys
	}
	return res
}

func buildGoogleBody(messages []ChatMessage, maxTokens int, temperature float64) map[string]any {
	contents := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		role := m.Role
		if role == "assistant" || role == "system" {
			role = "model"
		}
		contents = append(contents, map[string]any{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		})
	}
	return map[string]any{
		"contents": contents,
		"generationConfig": map[string]any{
			"temperature":     temperature,
			"maxOutputTokens": maxTokens,
		},
	}
}

// ParseProviderResponse extracts assistant text content across all supported provider schemas.
func ParseProviderResponse(provider string, data map[string]any) string {
	switch provider {
	case "google":
		cands, _ := data["candidates"].([]any)
		if len(cands) > 0 {
			c0, _ := cands[0].(map[string]any)
			content, _ := c0["content"].(map[string]any)
			parts, _ := content["parts"].([]any)
			if len(parts) > 0 {
				p0, _ := parts[0].(map[string]any)
				if s, ok := p0["text"].(string); ok {
					return s
				}
			}
		}
		return ""
	case "anthropic":
		items, _ := data["content"].([]any)
		if len(items) > 0 {
			c0, _ := items[0].(map[string]any)
			if s, ok := c0["text"].(string); ok {
				return s
			}
		}
		return ""
	default:
		choices, _ := data["choices"].([]any)
		if len(choices) > 0 {
			c0, _ := choices[0].(map[string]any)
			msg, _ := c0["message"].(map[string]any)
			if s, ok := msg["content"].(string); ok && s != "" {
				return s
			}
		}
		if msg, ok := data["message"].(map[string]any); ok {
			if s, okStr := msg["content"].(string); okStr && s != "" {
				return s
			}
		}
		if s, ok := data["response"].(string); ok {
			return s
		}
		return ""
	}
}

// ReadAPIError normalizes provider error JSON responses.
func ReadAPIError(status int, rawBody []byte, provider string) string {
	var parsed map[string]any
	if err := json.Unmarshal(rawBody, &parsed); err == nil {
		if errObj, ok := parsed["error"].(map[string]any); ok {
			if msg, okMsg := errObj["message"].(string); okMsg && msg != "" {
				return msg
			}
		}
		if msg, ok := parsed["message"].(string); ok && msg != "" {
			return msg
		}
	}
	return fmt.Sprintf("%s API error (HTTP %d)", provider, status)
}
