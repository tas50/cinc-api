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

func TestGenerateKeyPair(t *testing.T) {
	privPEM, pubPEM, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	// The private PEM must round-trip back through the library's own parser.
	priv, err := ParseKey([]byte(privPEM))
	if err != nil {
		t.Fatalf("ParseKey on generated private key: %v", err)
	}
	if bits := priv.N.BitLen(); bits != 2048 {
		t.Errorf("generated key is %d bits, want 2048", bits)
	}

	// The public PEM must be a PKIX RSA public key matching the private key.
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		t.Fatalf("public PEM block = %+v, want a PUBLIC KEY block", block)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("ParsePKIXPublicKey: %v", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("public key is %T, want *rsa.PublicKey", pub)
	}
	if rsaPub.N.Cmp(priv.N) != 0 || rsaPub.E != priv.E {
		t.Error("public key does not match the generated private key")
	}

	// Two calls must not produce identical key material.
	priv2, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair (second): %v", err)
	}
	if priv2 == privPEM {
		t.Error("two GenerateKeyPair calls returned identical private keys")
	}
}
