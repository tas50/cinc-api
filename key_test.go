// key_test.go
package cinc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestParseKey_PKCS1(t *testing.T) {
	pemBytes, err := os.ReadFile("testdata/test_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	key, err := ParseKey(pemBytes)
	if err != nil {
		t.Fatalf("ParseKey: %v", err)
	}
	if key == nil || key.N == nil {
		t.Fatal("expected a usable RSA key")
	}
}

func TestParseKey_PKCS8(t *testing.T) {
	// Re-wrap the PKCS#1 test key in PKCS#8 form so we can exercise that
	// branch of ParseKey without committing another fixture to disk.
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	got, err := ParseKey(pemBytes)
	if err != nil {
		t.Fatalf("ParseKey(PKCS#8): %v", err)
	}
	if got == nil || got.N.Cmp(rsaKey.N) != 0 {
		t.Fatal("parsed PKCS#8 key does not match generated key")
	}
}

func TestParseKey_PKCS8_NonRSA(t *testing.T) {
	// An EC private key wrapped in PKCS#8 should be rejected with a clear
	// "want *rsa.PrivateKey" message — we only support RSA.
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	_, err = ParseKey(pemBytes)
	if err == nil {
		t.Fatal("expected error parsing EC key, got nil")
	}
	if !contains(err.Error(), "rsa.PrivateKey") {
		t.Errorf("error %q should mention *rsa.PrivateKey", err.Error())
	}
}

func TestParseKey_NoPEMBlock(t *testing.T) {
	_, err := ParseKey([]byte("not a key"))
	if err == nil {
		t.Fatal("expected error for non-PEM input")
	}
	if !contains(err.Error(), "no PEM block") {
		t.Errorf("error %q should mention missing PEM block", err.Error())
	}
}

func TestParseKey_EmptyInput(t *testing.T) {
	if _, err := ParseKey(nil); err == nil {
		t.Fatal("expected error for nil input")
	}
	if _, err := ParseKey([]byte("")); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseKey_InvalidPEMBody(t *testing.T) {
	// Valid PEM envelope, but the bytes inside are not a private key.
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: []byte("garbage"),
	})
	if _, err := ParseKey(pemBytes); err == nil {
		t.Fatal("expected error for invalid PEM contents")
	}
}

func TestLoadKeyFile(t *testing.T) {
	key, err := LoadKeyFile("testdata/test_key.pem")
	if err != nil {
		t.Fatalf("LoadKeyFile: %v", err)
	}
	if key == nil {
		t.Fatal("expected a key")
	}
}

func TestLoadKeyFile_Missing(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.pem")
	_, err := LoadKeyFile(missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !contains(err.Error(), "read key file") {
		t.Errorf("error %q should mention reading the key file", err.Error())
	}
}
