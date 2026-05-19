// nodes_test.go
package cinc

import (
	"context"
	"net/http"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestNodes_CRUD(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/nodes/web01",
		cinctest.Route{Body: `{"name":"web01","chef_environment":"prod","run_list":["recipe[nginx]"],"automatic":{"fqdn":"web01.x"}}`})
	srv.Handle("POST /organizations/o/nodes",
		cinctest.Route{Status: 201, Body: `{"uri":"http://x/nodes/web02"}`,
			Assert: func(t *testing.T, r *http.Request, body []byte) {
				if !contains(string(body), "web02") {
					t.Errorf("create body missing name: %s", body)
				}
			}})
	srv.Handle("PUT /organizations/o/nodes/web01",
		cinctest.Route{Body: `{"name":"web01","chef_environment":"staging"}`})
	srv.Handle("DELETE /organizations/o/nodes/web01", cinctest.Route{Body: `{}`})
	srv.Handle("GET /organizations/o/nodes",
		cinctest.Route{Body: `{"web01":"http://x/nodes/web01"}`})

	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	n, _, err := c.Nodes.Get(ctx, "web01")
	if err != nil || n.Environment != "prod" ||
		n.Automatic.GetString("fqdn") != "web01.x" {
		t.Fatalf("Get: %+v %v", n, err)
	}
	if _, _, err := c.Nodes.Create(ctx, &Node{Name: "web02"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := c.Nodes.Update(ctx, &Node{Name: "web01", Environment: "staging"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := c.Nodes.Delete(ctx, "web01"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	names, _, err := c.Nodes.List(ctx)
	if err != nil || names["web01"] == "" {
		t.Fatalf("List: %+v %v", names, err)
	}
}
