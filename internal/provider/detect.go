package provider

import (
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// entry pairs a host pattern and optional path pattern with a factory function.
type entry struct {
	// hostSuffix is matched as a suffix of the request host (supports simple
	// wildcard prefix via leading "*").
	hostSuffix string
	// pathPrefix, if non-empty, additionally requires the path to start with it.
	pathPrefix string
	factory    func() Provider
}

// registry is evaluated in order; the first match wins.
var registry = []entry{
	{hostSuffix: "api.openai.com", factory: func() Provider { return &OpenAI{} }},
	{hostSuffix: "api.anthropic.com", factory: func() Provider { return &Anthropic{} }},
	// Google Gemini: dedicated parser handles "contents" format.
	{hostSuffix: "generativelanguage.googleapis.com", factory: func() Provider {
		return &Google{}
	}},
	// Mistral uses OpenAI-compatible format.
	{hostSuffix: "api.mistral.ai", factory: func() Provider {
		return NewGeneric(store.ProviderMistral)
	}},
	// Cohere uses OpenAI-compatible v2 format.
	{hostSuffix: "api.cohere.com", factory: func() Provider {
		return NewGeneric(store.ProviderCohere)
	}},
	// Groq/Together/Fireworks: OpenAI-compatible with streaming.
	{hostSuffix: "api.groq.com", factory: func() Provider {
		return NewGeneric(store.ProviderGroq)
	}},
	{hostSuffix: "api.together.xyz", factory: func() Provider {
		return NewGeneric(store.ProviderTogether)
	}},
	{hostSuffix: "api.fireworks.ai", factory: func() Provider {
		return NewGeneric(store.ProviderFireworks)
	}},
	// Ollama: dedicated parser for local model format.
	{hostSuffix: "localhost:11434", factory: func() Provider { return &Ollama{} }},
	{hostSuffix: "127.0.0.1:11434", factory: func() Provider { return &Ollama{} }},
	// OpenRouter: OpenAI-compatible with provider/model naming.
	{hostSuffix: "openrouter.ai", factory: func() Provider { return &OpenRouter{} }},
	// Azure OpenAI: dedicated parser extracts deployment name from URL.
	{hostSuffix: ".openai.azure.com", factory: func() Provider { return &Azure{} }},
	// AWS Bedrock: Generic (complex format, future dedicated parser).
	{hostSuffix: ".amazonaws.com", pathPrefix: "", factory: func() Provider {
		return NewGeneric(store.ProviderBedrock)
	}},
}

// Detect returns the appropriate Provider for the given host and path.
// It returns nil when the request is not directed at a known LLM endpoint
// and does not look like an OpenAI-compatible chat completions path.
func Detect(host, path string) Provider {
	hostLower := strings.ToLower(host)
	pathLower := strings.ToLower(path)

	for _, e := range registry {
		if matchHost(hostLower, e.hostSuffix) {
			if e.pathPrefix == "" || strings.HasPrefix(pathLower, e.pathPrefix) {
				return e.factory()
			}
		}
	}

	// Bedrock wildcard: bedrock-runtime.<region>.amazonaws.com
	if strings.HasPrefix(hostLower, "bedrock-runtime.") && strings.HasSuffix(hostLower, ".amazonaws.com") {
		return NewGeneric(store.ProviderBedrock)
	}

	// Fallback: OpenAI-compatible path.
	if strings.Contains(pathLower, "/v1/chat/completions") {
		return NewGeneric(store.ProviderCompatible)
	}

	return nil
}

// matchHost reports whether host matches the pattern.
// A pattern starting with "." is treated as a domain suffix (e.g. ".openai.azure.com"
// matches "myco.openai.azure.com"). Otherwise an exact match is required.
func matchHost(host, pattern string) bool {
	if strings.HasPrefix(pattern, ".") {
		return strings.HasSuffix(host, pattern)
	}
	return host == pattern
}
