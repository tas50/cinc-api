// client_test.go
package cinc

import (
	"strings"
	"testing"
	"time"
)

func TestNewClient_OK(t *testing.T) {
	c, err := NewClient(Config{
		ServerURL: "https://chef.example.com/", Org: "myorg",
		ClientName: "node1", Key: testRSAKey(t),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if got := c.orgPath("/nodes"); got != "/organizations/myorg/nodes" {
		t.Fatalf("orgPath = %q", got)
	}
	if c.Nodes == nil {
		t.Fatal("Nodes service not wired")
	}
}

func TestNewClient_BadConfig(t *testing.T) {
	if _, err := NewClient(Config{Org: "o"}); err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestNewClient_BadURL(t *testing.T) {
	if _, err := NewClient(Config{
		ServerURL: "://bad", Org: "o", ClientName: "c", Key: testRSAKey(t),
	}); err == nil {
		t.Fatal("expected error for malformed URL")
	}
}

func TestClient_now(t *testing.T) {
	c, _ := NewClient(Config{
		ServerURL: "https://h", Org: "o", ClientName: "c", Key: testRSAKey(t),
	})
	c.clock = func() time.Time { return time.Unix(0, 0).UTC() }
	if got := strings.TrimSpace(c.timestamp()); got != "1970-01-01T00:00:00Z" {
		t.Fatalf("timestamp = %q", got)
	}
}
