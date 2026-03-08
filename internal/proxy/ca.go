// Package proxy implements the MITM proxy server for intercepting LLM API calls.
package proxy

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// CA holds a local root CA certificate and private key used to
// dynamically sign per-hostname certificates for TLS interception.
type CA struct {
	cert    *x509.Certificate
	key     crypto.PrivateKey
	certPEM []byte
	keyPEM  []byte
}

// LoadOrCreateCA loads the CA from ~/.probe/ca-cert.pem and ~/.probe/ca-key.pem,
// creating new ones if they don't exist. The CA is a 10-year self-signed RSA-2048
// certificate with IsCA=true and KeyUsageCertSign.
func LoadOrCreateCA() (*CA, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ca: cannot determine home directory: %w", err)
	}

	probeDir := filepath.Join(home, ".probe")
	if err := os.MkdirAll(probeDir, 0700); err != nil {
		return nil, fmt.Errorf("ca: cannot create ~/.probe directory: %w", err)
	}

	certPath := filepath.Join(probeDir, "ca-cert.pem")
	keyPath := filepath.Join(probeDir, "ca-key.pem")

	certPEM, certErr := os.ReadFile(certPath)
	keyPEM, keyErr := os.ReadFile(keyPath)

	if certErr == nil && keyErr == nil {
		// Both files exist; try to parse them.
		ca, err := parseCAPEM(certPEM, keyPEM)
		if err == nil {
			return ca, nil
		}
		// Files are corrupt — fall through to regenerate.
	}

	// Generate a new CA.
	ca, err := generateCA()
	if err != nil {
		return nil, fmt.Errorf("ca: generating CA: %w", err)
	}

	if err := os.WriteFile(certPath, ca.certPEM, 0644); err != nil {
		return nil, fmt.Errorf("ca: writing %s: %w", certPath, err)
	}
	if err := os.WriteFile(keyPath, ca.keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("ca: writing %s: %w", keyPath, err)
	}

	return ca, nil
}

// CertPEM returns the PEM-encoded CA certificate.
func (ca *CA) CertPEM() []byte {
	return ca.certPEM
}

// PrintTrustInstructions prints OS-specific commands for trusting the CA
// certificate so that intercepted TLS connections appear valid.
func (ca *CA) PrintTrustInstructions() {
	home, _ := os.UserHomeDir()
	certPath := filepath.Join(home, ".probe", "ca-cert.pem")

	fmt.Println("To trust the probe CA certificate, run:")
	switch runtime.GOOS {
	case "darwin":
		fmt.Printf(
			"  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s\n",
			certPath,
		)
	case "linux":
		fmt.Printf("  sudo cp %s /usr/local/share/ca-certificates/probe-ca.crt\n", certPath)
		fmt.Println("  sudo update-ca-certificates")
	case "windows":
		fmt.Printf("  certutil -addstore -f ROOT %s\n", certPath)
	default:
		fmt.Printf("  Install %s into your system trust store.\n", certPath)
	}
}

// generateCA creates a new self-signed RSA-2048 CA certificate valid for 10 years.
func generateCA() (*CA, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating RSA key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("generating serial number: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "probe Local CA",
			Organization: []string{"probe"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("creating certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parsing generated certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshalling private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return &CA{
		cert:    cert,
		key:     key,
		certPEM: certPEM,
		keyPEM:  keyPEM,
	}, nil
}

// parseCAPEM parses PEM-encoded certificate and key bytes into a CA.
func parseCAPEM(certPEM, keyPEM []byte) (*CA, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("no PEM block in certificate file")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing certificate: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("no PEM block in key file")
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		// Try PKCS1 as fallback.
		rsaKey, err2 := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parsing private key (PKCS8: %v, PKCS1: %v)", err, err2)
		}
		key = rsaKey
	}

	return &CA{
		cert:    cert,
		key:     key,
		certPEM: certPEM,
		keyPEM:  keyPEM,
	}, nil
}
