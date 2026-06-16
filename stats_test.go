package cinc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStats_Get(t *testing.T) {
	// /_stats uses HTTP Basic auth, NOT the Chef signing protocol, so it is
	// served outside the cinctest harness and the request must NOT carry
	// X-Ops-Authorization headers.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_stats" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("format") != "json" {
			t.Errorf("format = %q", r.URL.Query().Get("format"))
		}
		if r.Header.Get("X-Ops-Authorization-1") != "" {
			t.Error("stats request must not be Chef-signed")
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "statsuser" || pass != "secret" {
			t.Errorf("basic auth = %q/%q ok=%v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"name":"erlang_vm_thread_pool_size","type":"GAUGE","help":"threads","metrics":[{"value":"5"}]},
			{"name":"pg_stat_seq_scan","type":"COUNTER","help":"scans","metrics":[{"value":"22147"}]}
		]`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	stats, resp, err := c.Stats.Get(context.Background(), "statsuser", "secret")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(stats) != 2 {
		t.Fatalf("got %d stats", len(stats))
	}
	if stats[0].Name != "erlang_vm_thread_pool_size" || stats[0].Type != "GAUGE" {
		t.Fatalf("stat[0] = %+v", stats[0])
	}
	if len(stats[0].Metrics) != 1 || stats[0].Metrics[0].Value != "5" {
		t.Fatalf("stat[0].Metrics = %+v", stats[0].Metrics)
	}
}

func TestStats_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	if _, _, err := c.Stats.Get(context.Background(), "statsuser", "secret"); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestStats_ConnError(t *testing.T) {
	// Point the client at a server that is already closed so the HTTP round
	// trip fails, exercising the transport error path.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	c := newTestClient(t, srv)
	srv.Close()
	if _, _, err := c.Stats.Get(context.Background(), "statsuser", "secret"); err == nil {
		t.Fatal("expected connection error")
	}
}

func TestStats_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":["unauthorized"]}`))
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	if _, _, err := c.Stats.Get(context.Background(), "statsuser", "wrong"); err == nil {
		t.Fatal("expected 401")
	}
}
