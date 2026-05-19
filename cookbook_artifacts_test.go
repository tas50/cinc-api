package cinc

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestCookbookArtifacts_GetAndList(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/cookbook_artifacts",
		cinctest.Route{Body: `{"nginx":{"url":"http://x","versions":[{"identifier":"abc123","url":"http://x/abc123"}]}}`})
	srv.Handle("GET /organizations/o/cookbook_artifacts/nginx/abc123",
		cinctest.Route{Body: `{"cookbook_name":"nginx","version":"abc123","name":"nginx"}`})
	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	list, _, err := c.CookbookArtifacts.List(ctx)
	if err != nil || len(list["nginx"].Versions) != 1 {
		t.Fatalf("List: %+v %v", list, err)
	}
	cb, _, err := c.CookbookArtifacts.Get(ctx, "nginx", "abc123")
	if err != nil || cb.CookbookName != "nginx" {
		t.Fatalf("Get: %+v %v", cb, err)
	}
}

func TestCookbookArtifacts_UploadRoundTrip(t *testing.T) {
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

	const identifier = "abc1234567890abcdef1234567890abcdef12345"

	// Pick recipes/default.rb as the one needing upload.
	recipeChecksum := md5Hex(recipeContent)

	var manifestBody []byte
	uploaded := map[string][]byte{}
	srv := cinctest.New(t)
	srv.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/organizations/o/sandboxes":
			uploadURL := "http://" + r.Host + "/upload/" + recipeChecksum
			w.WriteHeader(201)
			w.Write([]byte(`{"sandbox_id":"sb2","checksums":{"` + recipeChecksum + `":{"needs_upload":true,"url":"` + uploadURL + `"}}}`))
		case r.Method == "PUT" && r.URL.Path == "/upload/"+recipeChecksum:
			// Pre-signed bookshelf upload — must NOT carry Chef signing header.
			if r.Header.Get("X-Ops-Authorization-1") != "" {
				t.Errorf("file upload PUT carried Chef signing header (should be unsigned)")
			}
			body, _ := io.ReadAll(r.Body)
			uploaded[recipeChecksum] = body
			w.WriteHeader(200)
		case r.Method == "PUT" && r.URL.Path == "/organizations/o/sandboxes/sb2":
			w.Write([]byte(`{}`))
		case r.Method == "PUT" && r.URL.Path == "/organizations/o/cookbook_artifacts/nginx/"+identifier:
			manifestBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})

	c := newTestClient(t, srv.Server)
	if err := c.CookbookArtifacts.Upload(context.Background(), cb, identifier); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	// Issue 3: verify the file-upload PUT was exercised.
	got, ok := uploaded[recipeChecksum]
	if !ok {
		t.Fatal("expected recipes/default.rb to be uploaded, but no PUT was received")
	}
	if string(got) != string(recipeContent) {
		t.Errorf("uploaded recipes/default.rb = %q, want %q", got, recipeContent)
	}

	// Verify manifest was sent and contains identifier + correct chef_type.
	if len(manifestBody) == 0 {
		t.Fatal("no manifest body received")
	}
	body := string(manifestBody)
	if !contains(body, identifier) {
		t.Errorf("manifest missing identifier: %s", body)
	}
	if !contains(body, "cookbook_artifact_version") {
		t.Errorf("manifest missing chef_type cookbook_artifact_version: %s", body)
	}
}
