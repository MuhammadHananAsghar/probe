package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/analyze"
	"github.com/MuhammadHananAsghar/probe/internal/intercept"
	"github.com/MuhammadHananAsghar/probe/internal/provider"
	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/rs/zerolog"
)

// Config holds proxy server configuration.
type Config struct {
	// Port is the local port the proxy listens on.
	Port int
	// DashboardPort is the port for the web dashboard.
	DashboardPort int
	// StallThreshold is the duration after which a stream gap is flagged as a stall.
	StallThreshold time.Duration
	// NoBrowser disables automatic browser opening for the dashboard.
	NoBrowser bool
	// NoTLS disables TLS interception; only plain HTTP is proxied.
	NoTLS bool
	// Filter restricts interception to a provider or model (e.g. "anthropic" or "model=gpt-4o").
	Filter string
}

// providerHostMap maps provider names to their canonical API hostnames.
var providerHostMap = map[store.ProviderName]string{
	store.ProviderOpenAI:    "api.openai.com",
	store.ProviderAnthropic: "api.anthropic.com",
}

// Server is the probe proxy server. It intercepts HTTPS CONNECT tunnels
// destined for known LLM providers, performs TLS MITM, captures request/response
// data, and forwards traffic to the real upstream.
type Server struct {
	cfg       Config
	ca        *CA
	certCache *CertCache
	store     store.Store
	tracker   *analyze.Tracker
	eventCh   chan *store.Request
	mu        sync.Mutex
	log       *zerolog.Logger
}

// New creates a new proxy Server. It loads or creates the local CA and
// initialises the certificate cache.
func New(cfg Config, s store.Store, tracker *analyze.Tracker) (*Server, error) {
	ca, err := LoadOrCreateCA()
	if err != nil {
		return nil, fmt.Errorf("proxy: loading CA: %w", err)
	}

	l := zerolog.Nop()
	srv := &Server{
		cfg:       cfg,
		ca:        ca,
		certCache: NewCertCache(ca),
		store:     s,
		tracker:   tracker,
		eventCh:   make(chan *store.Request, 256),
		log:       &l,
	}

	return srv, nil
}

// SetLogger replaces the server's logger. Must be called before Start.
func (s *Server) SetLogger(l *zerolog.Logger) {
	s.log = l
}

// Start begins listening on cfg.Port. It blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(s.serveHTTP),
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("proxy: listen %s: %w", addr, err)
	}

	s.log.Info().Msgf("proxy listening on %s", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// Events returns a channel that receives a copy of each request as it completes.
func (s *Server) Events() <-chan *store.Request {
	return s.eventCh
}

// CA returns the proxy's local certificate authority.
func (s *Server) CA() *CA {
	return s.ca
}

// serveHTTP is the top-level HTTP handler.
func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.handleCONNECT(w, r)
		return
	}

	// Direct (non-CONNECT) request — base URL mode.
	p := s.detectProviderFromPath(r.Host, r.URL.Path)
	if p == nil {
		http.Error(w, "probe: no provider detected for this request", http.StatusBadGateway)
		return
	}

	if err := s.interceptRequest(w, r, p, r.URL.Host); err != nil {
		s.log.Error().Err(err).Msgf("intercept error for %s %s", r.Method, r.URL)
	}
}

// handleCONNECT processes HTTP CONNECT tunnel requests.
func (s *Server) handleCONNECT(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}

	if s.cfg.NoTLS || !isLLMHost(host) {
		if err := handlePassthrough(w, r); err != nil {
			s.log.Error().Err(err).Msgf("passthrough error for %s", host)
		}
		return
	}

	// TLS MITM for LLM hosts.
	if err := s.handleMITM(w, r, host); err != nil {
		s.log.Error().Err(err).Msgf("MITM error for %s", host)
	}
}

// handleMITM performs the TLS man-in-the-middle interception for the given host.
func (s *Server) handleMITM(w http.ResponseWriter, r *http.Request, hostWithPort string) error {
	hostname := hostWithPort
	if h, _, err := net.SplitHostPort(hostWithPort); err == nil {
		hostname = h
	}

	// Hijack the client connection.
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("MITM: ResponseWriter does not implement http.Hijacker")
	}

	rawConn, _, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("MITM: hijack: %w", err)
	}
	defer rawConn.Close()

	// Acknowledge the CONNECT.
	if _, err := fmt.Fprint(rawConn, "HTTP/1.1 200 Connection established\r\n\r\n"); err != nil {
		return fmt.Errorf("MITM: sending 200: %w", err)
	}

	// Wrap with TLS using a dynamically-signed certificate.
	tlsCfg, err := s.certCache.GetTLSConfig(hostname)
	if err != nil {
		return fmt.Errorf("MITM: getting TLS config for %s: %w", hostname, err)
	}

	tlsConn := tls.Server(rawConn, tlsCfg)
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("MITM: TLS handshake with client for %s: %w", hostname, err)
	}
	defer tlsConn.Close()

	// Serve HTTP/1.1 on the decrypted connection.
	providerHost := hostWithPort
	innerSrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Reconstruct absolute URL from the relative request inside the tunnel.
			if req.URL.Host == "" {
				req.URL.Host = providerHost
			}
			if req.URL.Scheme == "" {
				req.URL.Scheme = "https"
			}

			p := provider.Detect(hostname, req.URL.Path)
			if p == nil {
				// Unknown endpoint — passthrough without capture.
				if err := s.forwardRaw(w, req); err != nil {
					s.log.Error().Err(err).Msgf("forward error for %s", req.URL)
				}
				return
			}

			if err := s.interceptRequest(w, req, p, providerHost); err != nil {
				s.log.Error().Err(err).Msgf("intercept error for %s %s", req.Method, req.URL)
			}
		}),
	}

	httpConn := &singleConnListener{conn: tlsConn}
	_ = innerSrv.Serve(httpConn)
	return nil
}

// detectProviderFromPath maps a path (in base URL mode) to a provider.
// It handles the subset of paths needed for Phase 1: Anthropic /v1/messages
// and OpenAI /v1/chat/completions.
func (s *Server) detectProviderFromPath(host, path string) provider.Provider {
	switch {
	case strings.HasPrefix(path, "/v1/messages"):
		return provider.Detect("api.anthropic.com", path)
	case strings.HasPrefix(path, "/v1/chat/completions"),
		strings.HasPrefix(path, "/v1/completions"):
		return provider.Detect("api.openai.com", path)
	}
	return provider.Detect(host, path)
}

// interceptRequest captures, forwards, and records a single LLM API request.
func (s *Server) interceptRequest(w http.ResponseWriter, r *http.Request, p provider.Provider, upstreamHost string) error {
	// Determine the real upstream host to forward to.
	forwardHost := upstreamHost
	if mapped, ok := providerHostMap[p.Name()]; ok {
		forwardHost = mapped
	}

	// Build the store entry.
	req := &store.Request{
		ID:           newID(),
		StartedAt:    time.Now(),
		Method:       r.Method,
		URL:          r.URL.String(),
		Path:         r.URL.Path,
		Provider:     p.Name(),
		ProviderHost: forwardHost,
		Status:       store.StatusPending,
	}

	// Store request headers (masked).
	req.RequestHeaders = intercept.ExtractHeaders(r.Header)

	// Add to store (assigns Seq).
	seq := s.store.Add(req)
	req.Seq = seq

	// Drain the request body so we can both inspect and forward it.
	bodyBytes, err := intercept.DrainAndRestore(&r.Body)
	if err != nil {
		req.Status = store.StatusError
		req.ErrorMessage = fmt.Sprintf("reading request body: %v", err)
		s.store.Update(req)
		http.Error(w, req.ErrorMessage, http.StatusInternalServerError)
		return fmt.Errorf("reading body: %w", err)
	}
	req.RequestBody = bodyBytes

	// Parse the request for metadata.
	if parseErr := p.ParseRequest(bodyBytes, req); parseErr != nil {
		s.log.Warn().Err(parseErr).Msg("parsing request")
	}
	s.store.Update(req)

	// Build the forwarded request.
	targetURL := "https://" + forwardHost + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		req.Status = store.StatusError
		req.ErrorMessage = fmt.Sprintf("building upstream request: %v", err)
		s.store.Update(req)
		http.Error(w, req.ErrorMessage, http.StatusBadGateway)
		return fmt.Errorf("building upstream request: %w", err)
	}

	// Copy headers, excluding hop-by-hop and Host.
	copyHeaders(outReq.Header, r.Header)
	outReq.Header.Set("Host", forwardHost)
	outReq.Host = forwardHost

	// Forward the request.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, //nolint:gosec
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Do(outReq)
	if err != nil {
		req.Status = store.StatusError
		req.ErrorMessage = fmt.Sprintf("upstream request failed: %v", err)
		req.Latency = time.Since(req.StartedAt)
		s.store.Update(req)
		s.emit(req)
		http.Error(w, req.ErrorMessage, http.StatusBadGateway)
		return fmt.Errorf("upstream do: %w", err)
	}
	defer resp.Body.Close()

	req.StatusCode = resp.StatusCode
	req.ResponseHeaders = flattenHeaders(resp.Header)

	// Copy response headers to the client.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Capture the response body while streaming it to the client.
	var respBody []byte
	if req.Stream {
		req.Status = store.StatusStreaming
		s.store.Update(req)

		tee := intercept.NewTeeReadCloser(resp.Body)
		if _, err := io.Copy(w, tee); err != nil {
			s.log.Warn().Err(err).Msg("streaming copy interrupted")
			req.StreamStats = &store.StreamStats{Interrupted: true}
		}
		respBody = tee.Bytes()
	} else {
		var buf bytes.Buffer
		mw := io.MultiWriter(w, &buf)
		if _, err := io.Copy(mw, resp.Body); err != nil {
			s.log.Warn().Err(err).Msg("response copy error")
		}
		respBody = buf.Bytes()
	}

	req.ResponseBody = respBody
	req.Latency = time.Since(req.StartedAt)

	// Parse the response for tokens, finish reason, etc.
	if parseErr := p.ParseResponse(respBody, req); parseErr != nil {
		s.log.Warn().Err(parseErr).Msg("parsing response")
	}

	req.Status = store.StatusDone
	if resp.StatusCode >= 400 {
		req.Status = store.StatusError
	}
	req.EndedAt = time.Now()

	s.store.Update(req)
	s.tracker.Record(req)
	s.emit(req)

	return nil
}

// forwardRaw proxies a request to its upstream without any capture.
func (s *Server) forwardRaw(w http.ResponseWriter, r *http.Request) error {
	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return err
	}
	copyHeaders(outReq.Header, r.Header)

	resp, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

// emit sends a non-blocking copy of the request to the events channel.
func (s *Server) emit(req *store.Request) {
	select {
	case s.eventCh <- req:
	default:
		// Drop if consumers are not keeping up.
	}
}

// copyHeaders copies src headers into dst, skipping hop-by-hop headers.
func copyHeaders(dst, src http.Header) {
	hopByHop := map[string]bool{
		"Connection":          true,
		"Proxy-Connection":    true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}
	for k, vs := range src {
		if hopByHop[k] || strings.EqualFold(k, "host") {
			continue
		}
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

// flattenHeaders converts a http.Header into a simple map[string]string,
// joining multiple values with ", ".
func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, vs := range h {
		out[k] = strings.Join(vs, ", ")
	}
	return out
}

// newID generates a unique request ID using crypto/rand.
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based ID.
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	// Format as UUID-like string.
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}

// singleConnListener is a net.Listener that serves exactly one connection.
type singleConnListener struct {
	conn net.Conn
	once sync.Once
	ch   chan net.Conn
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	var c chan net.Conn
	l.once.Do(func() {
		l.ch = make(chan net.Conn, 1)
		l.ch <- l.conn
	})
	c = l.ch
	conn, ok := <-c
	if !ok {
		return nil, fmt.Errorf("listener closed")
	}
	close(c)
	return conn, nil
}

func (l *singleConnListener) Close() error {
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}

// Ensure singleConnListener implements net.Listener.
var _ net.Listener = (*singleConnListener)(nil)

// bufioConn wraps a net.Conn with a buffered reader (used when peeking at
// already-read bytes during connection hijacking).
type bufioConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *bufioConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}
