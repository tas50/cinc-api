// clients_test.go
package cinc

import (
	"context"
	"errors"
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

func TestClients_Update(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("PUT /organizations/o/clients/monkeypants",
		cinctest.Route{
			Body: `{"name":"monkeypants","clientname":"monkeypants","validator":true,"json_class":"Chef::ApiClient","chef_type":"client"}`,
		})

	c := newTestClient(t, srv.Server)
	ctx := context.Background()

	updated, _, err := c.Clients.Update(ctx, &APIClient{Name: "monkeypants", Validator: true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "monkeypants" {
		t.Fatalf("Update: expected name=monkeypants, got %q", updated.Name)
	}
	if !updated.Validator {
		t.Fatalf("Update: expected validator=true, got false")
	}
}

func TestClients_Update_Errors(t *testing.T) {
	t.Run("404_is_ErrNotFound", func(t *testing.T) {
		srv := cinctest.New(t)
		srv.Handle("PUT /organizations/o/clients/ghost",
			cinctest.Route{Status: 404, Body: `{"error":["client 'ghost' not found"]}`})
		c := newTestClient(t, srv.Server)
		_, _, err := c.Clients.Update(context.Background(), &APIClient{Name: "ghost"})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
