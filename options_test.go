// options_test.go
package cinc

import (
	"crypto/rsa"
	"os"
	"testing"
)

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	data, err := os.ReadFile("testdata/test_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	key, err := ParseKey(data)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestConfig_validate(t *testing.T) {
	good := Config{ServerURL: "https://h", Org: "o", ClientName: "c", Key: testRSAKey(t)}
	if err := good.validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	bad := Config{Org: "o", ClientName: "c", Key: testRSAKey(t)}
	if err := bad.validate(); err == nil {
		t.Fatal("expected error for missing ServerURL")
	}
}

func TestOptions_apply(t *testing.T) {
	o := defaultOptions()
	WithUserAgent("x/1")(&o)
	if o.userAgent != "x/1" {
		t.Fatalf("userAgent = %q", o.userAgent)
	}
}
