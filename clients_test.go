// clients_test.go
package cinc

import (
	"context"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestClients_CRUD(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/clients/node1",
		cinctest.Route{Body: `{"name":"node1","validator":false}`})
	srv.Handle("POST /organizations/o/clients",
		cinctest.Route{Status: 201,
			Body: `{"uri":"http://x/clients/node2","chef_key":{"private_key":"-----BEGIN"}}`})
	srv.Handle("DELETE /organizations/o/clients/node1", cinctest.Route{Body: `{}`})
	srv.Handle("GET /organizations/o/clients",
		cinctest.Route{Body: `{"node1":"http://x/clients/node1"}`})

	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	cl, _, err := c.Clients.Get(ctx, "node1")
	if err != nil || cl.Name != "node1" || cl.Validator {
		t.Fatalf("Get: %+v %v", cl, err)
	}
	created, _, err := c.Clients.Create(ctx, &APIClient{Name: "node2"})
	if err != nil || created.ChefKey.PrivateKey == "" {
		t.Fatalf("Create: %+v %v", created, err)
	}
	if _, err := c.Clients.Delete(ctx, "node1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if names, _, err := c.Clients.List(ctx); err != nil || names["node1"] == "" {
		t.Fatalf("List: %+v %v", names, err)
	}
}
