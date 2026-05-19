package cinc

import (
	"context"
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
