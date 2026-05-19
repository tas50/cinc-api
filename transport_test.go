// transport_test.go
package cinc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
