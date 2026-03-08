package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// llmHosts is the set of exact hostnames that belong to known LLM providers.
var llmHosts = map[string]bool{
	"api.openai.com":                     true,
	"api.anthropic.com":                  true,
	"generativelanguage.googleapis.com":  true,
	"api.mistral.ai":                     true,
	"api.cohere.com":                     true,
	"api.groq.com":                       true,
	"api.together.xyz":                   true,
	"api.fireworks.ai":                   true,
	"openrouter.ai":                      true,
}

// llmHostSuffixes is the set of domain suffixes whose traffic should be intercepted.
var llmHostSuffixes = []string{
	".openai.azure.com",
	".amazonaws.com",
}

// isLLMHost returns true if the given hostname belongs to a known LLM provider
// and should be intercepted rather than passed through.
func isLLMHost(host string) bool {
	// Strip port if present.
	h := host
	if idx := strings.LastIndex(h, ":"); idx != -1 {
		h = h[:idx]
	}

	if llmHosts[h] {
		return true
	}

	for _, suffix := range llmHostSuffixes {
		if strings.HasSuffix(h, suffix) {
			return true
		}
	}

	return false
}

// handlePassthrough tunnels a raw CONNECT connection to the target host
// without decrypting it. Used for non-LLM HTTPS traffic.
func handlePassthrough(w http.ResponseWriter, r *http.Request) error {
	target := r.Host
	if target == "" {
		target = r.URL.Host
	}

	// Dial the upstream server.
	upstreamConn, err := net.Dial("tcp", target)
	if err != nil {
		http.Error(w, fmt.Sprintf("passthrough: cannot connect to %s: %v", target, err), http.StatusBadGateway)
		return fmt.Errorf("passthrough: dial %s: %w", target, err)
	}
	defer upstreamConn.Close()

	// Hijack the client connection.
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("passthrough: ResponseWriter does not implement http.Hijacker")
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("passthrough: hijack: %w", err)
	}
	defer clientConn.Close()

	// Acknowledge the CONNECT request.
	if _, err := fmt.Fprint(clientConn, "HTTP/1.1 200 Connection established\r\n\r\n"); err != nil {
		return fmt.Errorf("passthrough: writing 200: %w", err)
	}

	// Bidirectional copy until either side closes.
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(upstreamConn, clientConn) //nolint:errcheck
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, upstreamConn) //nolint:errcheck
		done <- struct{}{}
	}()

	<-done
	return nil
}
