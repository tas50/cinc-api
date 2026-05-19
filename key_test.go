// key_test.go
package cinc

import (
	"os"
	"testing"
)

func TestParseKey_PKCS1(t *testing.T) {
	pem, err := os.ReadFile("testdata/test_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	key, err := ParseKey(pem)
	if err != nil {
		t.Fatalf("ParseKey: %v", err)
	}
	if key == nil || key.N == nil {
		t.Fatal("expected a usable RSA key")
	}
}

func TestParseKey_Invalid(t *testing.T) {
	if _, err := ParseKey([]byte("not a key")); err == nil {
		t.Fatal("expected error for invalid PEM")
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
