package signing

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

func benchKey(b *testing.B) *rsa.PrivateKey {
	b.Helper()
	data, err := os.ReadFile("../../testdata/test_key.pem")
	if err != nil {
		b.Fatal(err)
	}
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		b.Fatal(err)
	}
	return key
}

func benchmarkSignHeaders(b *testing.B, bodySize int) {
	key := benchKey(b)
	r := Request{
		Method: "POST", Path: "/organizations/o/nodes/web-prod-01",
		Body: make([]byte, bodySize), UserID: "client", Timestamp: "2024-01-01T00:00:00Z",
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := SignHeaders(r, key); err != nil {
			b.Fatal(err)
		}
	}
}

// SmallBody is the typical case (most requests have no or a tiny body); the RSA
// sign dominates. LargeBody exercises the body-hashing cost, which scaled 2x
// before the single-hash fix.
func BenchmarkSignHeaders_SmallBody(b *testing.B) { benchmarkSignHeaders(b, 64) }
func BenchmarkSignHeaders_LargeBody(b *testing.B) { benchmarkSignHeaders(b, 1<<20) }

func BenchmarkCanonicalPath(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = CanonicalPath("/organizations/myorg/nodes/web-prod-01")
	}
}
