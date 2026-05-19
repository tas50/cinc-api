// environments_test.go
package cinc

import (
	"context"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestEnvironments_CRUD(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/environments/prod",
		cinctest.Route{Body: `{"name":"prod","cookbook_versions":{"nginx":"= 1.2.0"}}`})
	srv.Handle("POST /organizations/o/environments",
		cinctest.Route{Status: 201, Body: `{"uri":"http://x/environments/dev"}`})
	srv.Handle("PUT /organizations/o/environments/prod",
		cinctest.Route{Body: `{"name":"prod"}`})
	srv.Handle("DELETE /organizations/o/environments/prod", cinctest.Route{Body: `{}`})
	srv.Handle("GET /organizations/o/environments",
		cinctest.Route{Body: `{"prod":"http://x/environments/prod"}`})

	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	e, _, err := c.Environments.Get(ctx, "prod")
	if err != nil || e.Name != "prod" || e.CookbookVersions["nginx"] != "= 1.2.0" {
		t.Fatalf("Get: %+v %v", e, err)
	}
	if _, _, err := c.Environments.Create(ctx, &Environment{Name: "dev"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := c.Environments.Update(ctx, &Environment{Name: "prod"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := c.Environments.Delete(ctx, "prod"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if names, _, err := c.Environments.List(ctx); err != nil || names["prod"] == "" {
		t.Fatalf("List: %+v %v", names, err)
	}
}
