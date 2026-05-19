// databags_test.go
package cinc

import (
	"context"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestDataBags(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/data",
		cinctest.Route{Body: `{"creds":"http://x/data/creds"}`})
	srv.Handle("POST /organizations/o/data",
		cinctest.Route{Status: 201, Body: `{"uri":"http://x/data/creds"}`})
	srv.Handle("DELETE /organizations/o/data/creds", cinctest.Route{Body: `{}`})
	srv.Handle("GET /organizations/o/data/creds",
		cinctest.Route{Body: `{"db":"http://x/data/creds/db"}`})
	srv.Handle("GET /organizations/o/data/creds/db",
		cinctest.Route{Body: `{"id":"db","password":"s3cret"}`})
	srv.Handle("POST /organizations/o/data/creds",
		cinctest.Route{Status: 201, Body: `{"id":"web"}`})
	srv.Handle("PUT /organizations/o/data/creds/db",
		cinctest.Route{Body: `{"id":"db","password":"new"}`})
	srv.Handle("DELETE /organizations/o/data/creds/db", cinctest.Route{Body: `{}`})

	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	if names, _, err := c.DataBags.List(ctx); err != nil || names["creds"] == "" {
		t.Fatalf("List: %+v %v", names, err)
	}
	if _, err := c.DataBags.Create(ctx, "creds"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	items := c.DataBags.Items("creds")
	if names, _, err := items.List(ctx); err != nil || names["db"] == "" {
		t.Fatalf("Items.List: %+v %v", names, err)
	}
	it, _, err := items.Get(ctx, "db")
	if err != nil || it["password"] != "s3cret" {
		t.Fatalf("Items.Get: %+v %v", it, err)
	}
	if _, _, err := items.Create(ctx, DataBagItem{"id": "web"}); err != nil {
		t.Fatalf("Items.Create: %v", err)
	}
	if _, _, err := items.Update(ctx, DataBagItem{"id": "db", "password": "new"}); err != nil {
		t.Fatalf("Items.Update: %v", err)
	}
	if _, err := items.Delete(ctx, "db"); err != nil {
		t.Fatalf("Items.Delete: %v", err)
	}
	if _, err := c.DataBags.Delete(ctx, "creds"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
