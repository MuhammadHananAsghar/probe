package provider

import (
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Azure implements Provider for Azure OpenAI Service.
// Azure uses the same request/response format as OpenAI, but the model
// (deployment name) is embedded in the URL path rather than the request body.
// URL pattern: /openai/deployments/{deployment}/chat/completions
type Azure struct {
	openai OpenAI
}

// Name returns the provider identifier.
func (a *Azure) Name() store.ProviderName { return store.ProviderAzureOpenAI }

// ParseRequest extracts the deployment name from req.Path and delegates to
// the OpenAI parser for the rest.
func (a *Azure) ParseRequest(body []byte, req *store.Request) error {
	// Extract deployment name from path: /openai/deployments/{name}/...
	if dep := extractAzureDeployment(req.Path); dep != "" {
		req.Model = dep
	}

	// Azure uses the "api-key" header; it is already captured in RequestHeaders.
	// Delegate to OpenAI parser for body parsing.
	if err := a.openai.ParseRequest(body, req); err != nil {
		return err
	}

	// If the OpenAI parser set a model from the body (unlikely for Azure), prefer
	// the deployment name we extracted.
	if dep := extractAzureDeployment(req.Path); dep != "" {
		req.Model = dep
	}

	return nil
}

// ParseResponse delegates to the OpenAI parser.
func (a *Azure) ParseResponse(body []byte, req *store.Request) error {
	return a.openai.ParseResponse(body, req)
}

// ParseEvent delegates to the OpenAI streaming parser.
func (a *Azure) ParseEvent(eventType, data string, req *store.Request) string {
	return a.openai.ParseEvent(eventType, data, req)
}

// extractAzureDeployment returns the deployment name from an Azure OpenAI URL path.
// Example: /openai/deployments/gpt-4o/chat/completions → "gpt-4o"
func extractAzureDeployment(path string) string {
	const prefix = "/openai/deployments/"
	lower := strings.ToLower(path)
	idx := strings.Index(lower, prefix)
	if idx < 0 {
		return ""
	}
	rest := path[idx+len(prefix):]
	end := strings.Index(rest, "/")
	if end < 0 {
		return rest
	}
	return rest[:end]
}
