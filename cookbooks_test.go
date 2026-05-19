package cinc

import (
	"context"
	"fmt"
	"io"
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

func TestCookbooks_Download(t *testing.T) {
	// The fake server doubles as: (a) the Chef API endpoint that returns the
	// manifest, and (b) the "bookshelf" that serves the pre-signed file URLs.
	srv := cinctest.New(t)
	srv.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/organizations/o/cookbooks/nginx/1.2.0":
			// Chef-signed request — must carry auth header.
			if r.Header.Get("X-Ops-Authorization-1") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			baseURL := "http://" + r.Host
			manifest := fmt.Sprintf(`{
				"cookbook_name":"nginx","name":"nginx-1.2.0","version":"1.2.0",
				"recipes":[{"name":"default.rb","path":"recipes/default.rb","specificity":"default","checksum":"abc","url":"%s/files/recipes/default.rb"}],
				"root_files":[{"name":"metadata.rb","path":"metadata.rb","specificity":"default","checksum":"def","url":"%s/files/metadata.rb"}]
			}`, baseURL, baseURL)
			w.Write([]byte(manifest))
		case r.Method == "GET" && r.URL.Path == "/files/recipes/default.rb":
			// Bookshelf: plain unsigned request.
			w.Write([]byte("package 'nginx'\n"))
		case r.Method == "GET" && r.URL.Path == "/files/metadata.rb":
			w.Write([]byte("name 'nginx'\nversion '1.2.0'\n"))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotImplemented)
		}
	})

	c := newTestClient(t, srv.Server)
	dest := t.TempDir()
	if err := c.Cookbooks.Download(context.Background(), "nginx", "1.2.0", dest); err != nil {
		t.Fatalf("Download: %v", err)
	}

	recipe := filepath.Join(dest, "recipes", "default.rb")
	if data, err := os.ReadFile(recipe); err != nil || string(data) != "package 'nginx'\n" {
		t.Fatalf("recipe file: data=%q err=%v", data, err)
	}
	metadata := filepath.Join(dest, "metadata.rb")
	if data, err := os.ReadFile(metadata); err != nil || string(data) != "name 'nginx'\nversion '1.2.0'\n" {
		t.Fatalf("metadata file: data=%q err=%v", data, err)
	}
}

func TestCookbooks_Download_PathTraversal(t *testing.T) {
	// A malicious server returns a manifest with a path that escapes destDir.
	srv := cinctest.New(t)
	srv.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/organizations/o/cookbooks/evil/1.0.0":
			if r.Header.Get("X-Ops-Authorization-1") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			baseURL := "http://" + r.Host
			// The path component traverses up out of destDir.
			manifest := fmt.Sprintf(`{
				"cookbook_name":"evil","name":"evil-1.0.0","version":"1.0.0",
				"root_files":[{"name":"escape.txt","path":"../escape.txt","specificity":"default","checksum":"aaa","url":"%s/files/escape.txt"}]
			}`, baseURL)
			w.Write([]byte(manifest))
		case r.Method == "GET" && r.URL.Path == "/files/escape.txt":
			w.Write([]byte("pwned\n"))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotImplemented)
		}
	})

	c := newTestClient(t, srv.Server)
	dest := t.TempDir()
	err := c.Cookbooks.Download(context.Background(), "evil", "1.0.0", dest)
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
	// The escaped file must NOT have been written.
	escapedPath := filepath.Join(filepath.Dir(dest), "escape.txt")
	if _, statErr := os.Stat(escapedPath); statErr == nil {
		t.Errorf("path traversal: file was written outside destDir at %s", escapedPath)
	}
}

func TestCookbooks_UploadRoundTrip(t *testing.T) {
	// Build a cookbook on disk: metadata.rb + recipes/default.rb.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "nginx", "recipes"), 0o755)
	metadataContent := []byte("name 'nginx'\nversion '1.2.0'\n")
	recipeContent := []byte("package 'nginx'\n")
	os.WriteFile(filepath.Join(dir, "nginx", "metadata.rb"), metadataContent, 0o644)
	os.WriteFile(filepath.Join(dir, "nginx", "recipes", "default.rb"), recipeContent, 0o644)

	cb, err := cookbookFromDir(filepath.Join(dir, "nginx"), "1.2.0")
	if err != nil {
		t.Fatalf("cookbookFromDir: %v", err)
	}
	if len(cb.files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(cb.files))
	}

	// Pick the checksum of metadata.rb as the one needing upload.
	metadataChecksum := md5Hex(metadataContent)

	// Stateful fake: sandbox -> file PUT -> manifest PUT.
	srv := cinctest.New(t)
	uploaded := map[string][]byte{}
	var manifestBody []byte
	srv.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/organizations/o/sandboxes":
			// Return metadata.rb as needing upload; recipes/default.rb does not.
			uploadURL := "http://" + r.Host + "/upload/" + metadataChecksum
			w.WriteHeader(201)
			w.Write([]byte(`{"sandbox_id":"sb1","checksums":{"` + metadataChecksum + `":{"needs_upload":true,"url":"` + uploadURL + `"}}}`))
		case r.Method == "PUT" && r.URL.Path == "/upload/"+metadataChecksum:
			// Pre-signed bookshelf upload — must NOT carry Chef signing header.
			if r.Header.Get("X-Ops-Authorization-1") != "" {
				t.Errorf("file upload PUT carried Chef signing header (should be unsigned)")
			}
			body, _ := io.ReadAll(r.Body)
			uploaded[metadataChecksum] = body
			w.WriteHeader(200)
		case r.Method == "PUT" && r.URL.Path == "/organizations/o/sandboxes/sb1":
			w.Write([]byte(`{}`))
		case r.Method == "PUT" && r.URL.Path == "/organizations/o/cookbooks/nginx/1.2.0":
			manifestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(201)
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})
	c := newTestClient(t, srv.Server)
	if err := c.Cookbooks.Upload(context.Background(), cb); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	// Issue 3: verify the file-upload PUT was exercised.
	got, ok := uploaded[metadataChecksum]
	if !ok {
		t.Fatal("expected metadata.rb to be uploaded, but no PUT was received")
	}
	if string(got) != string(metadataContent) {
		t.Errorf("uploaded metadata.rb = %q, want %q", got, metadataContent)
	}

	// Issue 2: verify manifest includes chef_type "cookbook_version".
	if len(manifestBody) == 0 {
		t.Fatal("no manifest body received")
	}
	body := string(manifestBody)
	if !contains(body, "cookbook_version") {
		t.Errorf("manifest missing chef_type=cookbook_version: %s", body)
	}
}
