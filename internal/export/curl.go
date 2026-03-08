package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// apiKeyEnvVars maps provider names to their conventional API key environment
// variable names. The actual key value is never emitted.
var apiKeyEnvVars = map[store.ProviderName]string{
	store.ProviderOpenAI:      "OPENAI_API_KEY",
	store.ProviderAnthropic:   "ANTHROPIC_API_KEY",
	store.ProviderGoogle:      "GOOGLE_API_KEY",
	store.ProviderMistral:     "MISTRAL_API_KEY",
	store.ProviderCohere:      "COHERE_API_KEY",
	store.ProviderGroq:        "GROQ_API_KEY",
	store.ProviderTogether:    "TOGETHER_API_KEY",
	store.ProviderFireworks:   "FIREWORKS_API_KEY",
	store.ProviderOllama:      "",
	store.ProviderOpenRouter:  "OPENROUTER_API_KEY",
	store.ProviderAzureOpenAI: "AZURE_OPENAI_API_KEY",
	store.ProviderBedrock:     "AWS_ACCESS_KEY_ID",
	store.ProviderCompatible:  "API_KEY",
}

// sensitiveHeaders is the set of header names (lowercase) whose values are
// always masked in curl output.
var sensitiveHeaders = map[string]bool{
	"authorization":   true,
	"x-api-key":       true,
	"x-goog-api-key":  true,
	"api-key":         true,
}

// ToCurl generates a copy-pasteable curl command that reproduces the request.
// API keys are replaced with $ENV_VAR placeholders — the actual key is never
// emitted. The lang parameter selects output format: "curl" (default),
// "python", or "node".
func ToCurl(req *store.Request, lang string) string {
	switch lang {
	case "python":
		return toPython(req)
	case "node":
		return toNode(req)
	default:
		return toCurlDefault(req)
	}
}

func toCurlDefault(req *store.Request) string {
	var sb strings.Builder

	// Stream flag
	streamFlag := ""
	if req.Stream {
		streamFlag = " \\\n  -N"
	}

	sb.WriteString(fmt.Sprintf("curl -s -X %s \\\n  '%s'%s", req.Method, req.URL, streamFlag))

	// Headers
	for k, v := range req.RequestHeaders {
		lower := strings.ToLower(k)
		if sensitiveHeaders[lower] {
			envVar := maskHeader(k, req.Provider)
			sb.WriteString(fmt.Sprintf(" \\\n  -H '%s: $%s'", k, envVar))
		} else if lower != "content-length" && lower != "transfer-encoding" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H '%s: %s'", k, v))
		}
	}

	// Body
	if len(req.RequestBody) > 0 {
		// Pretty-print JSON body for readability.
		var body any
		if json.Unmarshal(req.RequestBody, &body) == nil {
			pretty, _ := json.MarshalIndent(body, "  ", "  ")
			sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", escapeSingle(string(pretty))))
		} else {
			sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", escapeSingle(string(req.RequestBody))))
		}
	}

	return sb.String()
}

func toPython(req *store.Request) string {
	var sb strings.Builder
	sb.WriteString("import requests\n\n")
	sb.WriteString("headers = {\n")
	for k, v := range req.RequestHeaders {
		lower := strings.ToLower(k)
		if sensitiveHeaders[lower] {
			env := maskHeader(k, req.Provider)
			sb.WriteString(fmt.Sprintf("    %q: f\"${%s}\",\n", k, env))
		} else if lower != "content-length" && lower != "transfer-encoding" {
			sb.WriteString(fmt.Sprintf("    %q: %q,\n", k, v))
		}
	}
	sb.WriteString("}\n\n")

	if len(req.RequestBody) > 0 {
		var body any
		if json.Unmarshal(req.RequestBody, &body) == nil {
			pretty, _ := json.MarshalIndent(body, "", "    ")
			sb.WriteString(fmt.Sprintf("data = %s\n\n", string(pretty)))
			sb.WriteString(fmt.Sprintf("resp = requests.%s(%q, headers=headers, json=data)\n",
				strings.ToLower(req.Method), req.URL))
		} else {
			sb.WriteString(fmt.Sprintf("resp = requests.%s(%q, headers=headers, data=%q)\n",
				strings.ToLower(req.Method), req.URL, string(req.RequestBody)))
		}
	} else {
		sb.WriteString(fmt.Sprintf("resp = requests.%s(%q, headers=headers)\n",
			strings.ToLower(req.Method), req.URL))
	}
	sb.WriteString("print(resp.json())\n")
	return sb.String()
}

func toNode(req *store.Request) string {
	var sb strings.Builder
	sb.WriteString("const response = await fetch(")
	sb.WriteString(fmt.Sprintf("%q, {\n", req.URL))
	sb.WriteString(fmt.Sprintf("  method: %q,\n", req.Method))
	sb.WriteString("  headers: {\n")
	for k, v := range req.RequestHeaders {
		lower := strings.ToLower(k)
		if sensitiveHeaders[lower] {
			env := maskHeader(k, req.Provider)
			sb.WriteString(fmt.Sprintf("    %q: process.env.%s,\n", k, env))
		} else if lower != "content-length" && lower != "transfer-encoding" {
			sb.WriteString(fmt.Sprintf("    %q: %q,\n", k, v))
		}
	}
	sb.WriteString("  },\n")
	if len(req.RequestBody) > 0 {
		sb.WriteString(fmt.Sprintf("  body: JSON.stringify(%s),\n", string(req.RequestBody)))
	}
	sb.WriteString("});\n")
	sb.WriteString("const data = await response.json();\n")
	sb.WriteString("console.log(data);\n")
	return sb.String()
}

// maskHeader returns the env-var placeholder name for a sensitive header.
func maskHeader(header string, provider store.ProviderName) string {
	lower := strings.ToLower(header)
	if lower == "authorization" || lower == "x-api-key" || lower == "api-key" {
		if env, ok := apiKeyEnvVars[provider]; ok && env != "" {
			return env
		}
		return "API_KEY"
	}
	// Convert header name to screaming snake case.
	return strings.ToUpper(strings.ReplaceAll(header, "-", "_"))
}

// escapeSingle escapes single quotes inside a shell single-quoted string by
// ending the quote, inserting an escaped single quote, and reopening.
func escapeSingle(s string) string {
	return strings.ReplaceAll(s, "'", `'\''`)
}
