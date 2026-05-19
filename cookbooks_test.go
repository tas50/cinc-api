package cinc

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestCookbooks_GetAndList(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/cookbooks",
		cinctest.Route{Body: `{"nginx":{"url":"http://x/cookbooks/nginx","versions":[{"version":"1.2.0","url":"http://x/cookbooks/nginx/1.2.0"}]}}`})
	srv.Handle("GET /organizations/o/cookbooks/nginx/1.2.0",
		cinctest.Route{Body: `{"cookbook_name":"nginx","version":"1.2.0","name":"nginx-1.2.0"}`})
	srv.Handle("DELETE /organizations/o/cookbooks/nginx/1.2.0",
		cinctest.Route{Body: `{}`})

	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	list, _, err := c.Cookbooks.List(ctx)
	if err != nil || len(list["nginx"].Versions) != 1 {
		t.Fatalf("List: %+v %v", list, err)
	}
	cb, _, err := c.Cookbooks.Get(ctx, "nginx", "1.2.0")
	if err != nil || cb.Version != "1.2.0" {
		t.Fatalf("Get: %+v %v", cb, err)
	}
	if _, err := c.Cookbooks.Delete(ctx, "nginx", "1.2.0"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestCookbooks_UploadRoundTrip(t *testing.T) {
	// Build a cookbook on disk: metadata.rb + recipes/default.rb.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "nginx", "recipes"), 0o755)
	os.WriteFile(filepath.Join(dir, "nginx", "metadata.rb"),
		[]byte("name 'nginx'\nversion '1.2.0'\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "nginx", "recipes", "default.rb"),
		[]byte("package 'nginx'\n"), 0o644)

	cb, err := cookbookFromDir(filepath.Join(dir, "nginx"), "1.2.0")
	if err != nil {
		t.Fatalf("cookbookFromDir: %v", err)
	}
	if len(cb.files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(cb.files))
	}

	// Stateful fake: sandbox -> file PUT -> manifest PUT.
	srv := cinctest.New(t)
	uploaded := map[string][]byte{}
	srv.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/organizations/o/sandboxes":
			w.WriteHeader(201)
			w.Write([]byte(`{"sandbox_id":"sb1","checksums":{}}`))
		case r.Method == "PUT" && r.URL.Path == "/organizations/o/sandboxes/sb1":
			w.Write([]byte(`{}`))
		case r.Method == "PUT" && r.URL.Path == "/organizations/o/cookbooks/nginx/1.2.0":
			w.WriteHeader(201)
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_ = uploaded
	})
	c := newTestClient(t, srv.Server)
	if err := c.Cookbooks.Upload(context.Background(), cb); err != nil {
		t.Fatalf("Upload: %v", err)
	}
}
