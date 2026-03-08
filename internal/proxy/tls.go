package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// CertCache caches dynamically-generated TLS configs per hostname.
// It generates each cert lazily and caches the result.
type CertCache struct {
	ca    *CA
	mu    sync.RWMutex
	cache map[string]*tls.Config
}

// NewCertCache creates a new CertCache backed by the given CA.
func NewCertCache(ca *CA) *CertCache {
	return &CertCache{
		ca:    ca,
		cache: make(map[string]*tls.Config),
	}
}

// GetTLSConfig returns (or generates and caches) a *tls.Config for the given hostname.
// The generated certificate is signed by the CA, valid for 1 year, covering the exact hostname.
func (cc *CertCache) GetTLSConfig(hostname string) (*tls.Config, error) {
	// Fast path: check cache with read lock.
	cc.mu.RLock()
	if cfg, ok := cc.cache[hostname]; ok {
		cc.mu.RUnlock()
		return cfg, nil
	}
	cc.mu.RUnlock()

	// Slow path: generate cert and cache it.
	cfg, err := cc.generateTLSConfig(hostname)
	if err != nil {
		return nil, err
	}

	cc.mu.Lock()
	cc.cache[hostname] = cfg
	cc.mu.Unlock()

	return cfg, nil
}

// generateTLSConfig creates a new TLS config with a freshly signed cert for hostname.
func (cc *CertCache) generateTLSConfig(hostname string) (*tls.Config, error) {
	// Generate RSA-2048 leaf key.
	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("tls: generating key for %s: %w", hostname, err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("tls: generating serial for %s: %w", hostname, err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: hostname,
		},
		DNSNames:  []string{hostname},
		NotBefore: now.Add(-time.Minute), // slight back-date to handle clock skew
		NotAfter:  now.Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	caKey, ok := cc.ca.key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("tls: CA key is not RSA")
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, cc.ca.cert, &leafKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("tls: signing cert for %s: %w", hostname, err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(leafKey)
	if err != nil {
		return nil, fmt.Errorf("tls: marshalling key for %s: %w", hostname, err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("tls: building tls.Certificate for %s: %w", hostname, err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}

	return cfg, nil
}
