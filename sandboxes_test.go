package cinc

import (
	"context"
	"net/http"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestSandbox_CreateAndUpload(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("POST /organizations/o/sandboxes",
		cinctest.Route{Status: 201, Body: `{
			"sandbox_id":"sb1",
			"checksums":{
				"` + md5Hex([]byte("hello")) + `":{"needs_upload":true,"url":"` + "PUTURL" + `"}
			}}`})
	c := newTestClient(t, srv.Server)

	sb, _, err := c.createSandbox(context.Background(),
		[]string{md5Hex([]byte("hello"))})
	if err != nil {
		t.Fatalf("createSandbox: %v", err)
	}
	entry := sb.Checksums[md5Hex([]byte("hello"))]
	if !entry.NeedsUpload || entry.URL == "" {
		t.Fatalf("checksum entry: %+v", entry)
	}

	// uploadFile PUTs raw bytes to the signed URL.
	put := cinctest.New(t)
	var gotBody string
	put.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		r.Body.Read(b)
		gotBody = string(b)
		w.WriteHeader(200)
	})
	if err := c.uploadFile(context.Background(), put.Server.URL, []byte("hello")); err != nil {
		t.Fatalf("uploadFile: %v", err)
	}
	if gotBody != "hello" {
		t.Fatalf("uploaded body = %q", gotBody)
	}
}
