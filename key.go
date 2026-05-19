// Package cinc is a Go client for the Chef/CINC Server API.
package cinc

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// ParseKey parses a PEM-encoded RSA private key (PKCS#1 or PKCS#8).
func ParseKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("cinc: no PEM block found in key data")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cinc: parse private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("cinc: key is %T, want *rsa.PrivateKey", parsed)
	}
	return key, nil
}

// LoadKeyFile reads and parses an RSA private key from a file path.
func LoadKeyFile(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cinc: read key file: %w", err)
	}
	return ParseKey(data)
}
