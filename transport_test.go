// transport_test.go
package cinc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tas50/cinc-api/internal/signing"
)

// collectAuthSig reassembles the X-Ops-Authorization-N chunks from an HTTP
// request into a single base64 string.
func collectAuthSig(r *http.Request) string {
	var b strings.Builder
	for i := 1; ; i++ {
		chunk := r.Header.Get("X-Ops-Authorization-" + strconv.Itoa(i))
		if chunk == "" {
			break
		}
		b.WriteString(chunk)
	}
	return b.String()
}

// TestTransport_SigningExcludesQueryString verifies that when a request path
// contains a query string the signing canonical path is stripped of the "?"
// and everything after it — i.e. the v1.3 spec requirement.
func TestTransport_SigningExcludesQueryString(t *testing.T) {
	const pathOnly = "/organizations/o/search/node"
	const qs = "q=name%3Aweb01&rows=10&start=0"
	var (
		capturedSig       string
		capturedTimestamp string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSig = collectAuthSig(r)
		capturedTimestamp = r.Header.Get("X-Ops-Timestamp")
		w.Write([]byte(`{"total":0,"start":0,"rows":[]}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	if _, _, err := do[SearchResult](context.Background(), c, "GET",
		pathOnly+"?"+qs, nil); err != nil {
		t.Fatalf("do: %v", err)
	}

	// Verify the signature against a canonical request built with the path
	// ONLY (no query string).  If the bug is present the signature was over
	// path+qs and this will fail.
	sigBytes, err := base64.StdEncoding.DecodeString(capturedSig)
	if err != nil {
		t.Fatalf("decode sig: %v", err)
	}
	canonical := signing.CanonicalRequest(signing.Request{
		Method:    "GET",
		Path:      pathOnly,
		Body:      nil,
		UserID:    "c",
		Timestamp: capturedTimestamp,
	})
	digest := sha256.Sum256([]byte(canonical))
	key := testRSAKey(t)
	if err := rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, digest[:], sigBytes); err != nil {
		t.Errorf("signature does not verify against path-only canonical request: %v\n"+
			"(this means the query string was included in the signed path)", err)
	}
}

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(Config{
		ServerURL: srv.URL, Org: "o", ClientName: "c", Key: testRSAKey(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDo_DecodesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Ops-Authorization-1") == "" {
			t.Error("request was not signed")
		}
		w.Write([]byte(`{"name":"web01"}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	type obj struct{ Name string }
	got, resp, err := do[obj](context.Background(), c, "GET", "/nodes/web01", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "web01" || resp.StatusCode != 200 {
		t.Fatalf("got %+v status %d", got, resp.StatusCode)
	}
}

func TestDo_MapsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":["no node"]}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	type obj struct{ Name string }
	_, _, err := do[obj](context.Background(), c, "GET", "/nodes/x", nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	c := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	type obj struct{}
	if _, _, err := do[obj](ctx, c, "GET", "/x", nil); err == nil {
		t.Fatal("expected context error")
	}
}

// TestRetry_503ThenSuccess asserts that a GET retried on 503 eventually
// succeeds and that the server is hit exactly 3 times (2 failures + 1 success).
func TestRetry_503ThenSuccess(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"name":"ok"}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	type obj struct{ Name string }
	got, _, err := do[obj](context.Background(), c, "GET", "/nodes/x", nil)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if got.Name != "ok" {
		t.Fatalf("unexpected body: %+v", got)
	}
	if n := hits.Load(); n != 3 {
		t.Fatalf("server hit %d times, want 3", n)
	}
}

// TestRetry_ContextCancelledNoRetry asserts that a cancelled context causes
// exactly one attempt (no retries) — validates the isNetErr fix for Issue 3.
func TestRetry_ContextCancelledNoRetry(t *testing.T) {
	var hits atomic.Int32
	// Use a channel to synchronise: the handler blocks until the client
	// has cancelled the context, ensuring the cancellation is observed.
	ready := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		close(ready) // signal we're in the handler
		// Block until the request context is cancelled (client gone).
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		type obj struct{}
		_, _, err := do[obj](ctx, c, "GET", "/nodes/x", nil)
		done <- err
	}()

	<-ready  // wait until the handler is entered
	cancel() // cancel only after the first request is in-flight

	err := <-done
	if err == nil {
		t.Fatal("expected error after context cancel")
	}
	// Give any spurious retry a moment to arrive, then check.
	if n := hits.Load(); n != 1 {
		t.Fatalf("server hit %d times after context cancel, want 1 (no retries)", n)
	}
}

// TestRetry_PostNot retried asserts that a non-GET (POST) 503 is never retried.
func TestRetry_PostNotRetried(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	type obj struct{}
	_, _, err := do[obj](context.Background(), c, "POST", "/nodes", map[string]any{"name": "x"})
	if err == nil {
		t.Fatal("expected error for 503")
	}
	if n := hits.Load(); n != 1 {
		t.Fatalf("POST hit %d times, want exactly 1 (must not retry non-GET)", n)
	}
}
