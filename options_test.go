// options_test.go
package cinc

import (
	"crypto/rsa"
	"net/http"
	"os"
	"testing"
	"time"
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
	key := testRSAKey(t)
	cases := []struct {
		name    string
		cfg     Config
		wantErr string // substring expected in error message; "" means must succeed
	}{
		{"ok", Config{ServerURL: "https://h", Org: "o", ClientName: "c", Key: key}, ""},
		{"missing ServerURL", Config{Org: "o", ClientName: "c", Key: key}, "ServerURL"},
		{"missing Org", Config{ServerURL: "https://h", ClientName: "c", Key: key}, "Org"},
		{"missing ClientName", Config{ServerURL: "https://h", Org: "o", Key: key}, "ClientName"},
		{"missing Key", Config{ServerURL: "https://h", Org: "o", ClientName: "c"}, "Key"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestOptions_apply(t *testing.T) {
	o := defaultOptions()
	WithUserAgent("x/1")(&o)
	if o.userAgent != "x/1" {
		t.Fatalf("userAgent = %q", o.userAgent)
	}
}

func TestWithHTTPClient_Sets(t *testing.T) {
	o := defaultOptions()
	custom := &http.Client{Timeout: 7 * time.Second}
	WithHTTPClient(custom)(&o)
	if o.httpClient != custom {
		t.Fatalf("httpClient not replaced")
	}
}

func TestWithHTTPClient_NilIgnored(t *testing.T) {
	o := defaultOptions()
	original := o.httpClient
	WithHTTPClient(nil)(&o)
	if o.httpClient != original {
		t.Fatalf("nil http.Client should be ignored, got %v", o.httpClient)
	}
}

func TestWithChefVersion(t *testing.T) {
	o := defaultOptions()
	WithChefVersion("18.7.5")(&o)
	if o.chefVersion != "18.7.5" {
		t.Fatalf("chefVersion = %q, want 18.7.5", o.chefVersion)
	}
}

func TestWithSkipTLSVerify(t *testing.T) {
	o := defaultOptions()
	if o.skipTLSVerify {
		t.Fatal("default skipTLSVerify should be false")
	}
	WithSkipTLSVerify(true)(&o)
	if !o.skipTLSVerify {
		t.Fatal("WithSkipTLSVerify(true) did not enable the flag")
	}
}

func TestWithMaxRetries_Sets(t *testing.T) {
	o := defaultOptions()
	WithMaxRetries(5)(&o)
	if o.maxRetries != 5 {
		t.Fatalf("maxRetries = %d, want 5", o.maxRetries)
	}
}

func TestWithMaxRetries_NegativeIgnored(t *testing.T) {
	o := defaultOptions()
	original := o.maxRetries
	WithMaxRetries(-1)(&o)
	if o.maxRetries != original {
		t.Fatalf("negative maxRetries should be ignored: got %d, want %d", o.maxRetries, original)
	}
}

func TestWithMaxRetries_Zero(t *testing.T) {
	o := defaultOptions()
	WithMaxRetries(0)(&o)
	if o.maxRetries != 0 {
		t.Fatalf("WithMaxRetries(0) should disable retries; got %d", o.maxRetries)
	}
}

func TestDefaultOptions(t *testing.T) {
	o := defaultOptions()
	if o.httpClient == nil {
		t.Fatal("default httpClient is nil")
	}
	if o.httpClient.Timeout != 30*time.Second {
		t.Errorf("default Timeout = %v, want 30s", o.httpClient.Timeout)
	}
	if o.userAgent == "" {
		t.Error("default userAgent is empty")
	}
	if o.chefVersion == "" {
		t.Error("default chefVersion is empty")
	}
	if o.maxRetries < 1 {
		t.Errorf("default maxRetries = %d, want >= 1", o.maxRetries)
	}
}
