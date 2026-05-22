package cinc

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestACLs_Get(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/nodes/web01/_acl", cinctest.Route{
		Body: `{
			"create":{"actors":["pivotal"],"groups":["admins"]},
			"read":{"actors":[],"groups":["admins","users","clients"]},
			"update":{"actors":["pivotal"],"groups":["admins"]},
			"delete":{"actors":["pivotal"],"groups":["admins"]},
			"grant":{"actors":["pivotal"],"groups":["admins"]}
		}`,
	})
	c := newTestClient(t, srv.Server)

	acl, _, err := c.ACLs.Get(context.Background(), "nodes", "web01")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !reflect.DeepEqual(acl.Create.Actors, []string{"pivotal"}) {
		t.Errorf("create.actors = %v", acl.Create.Actors)
	}
	if !reflect.DeepEqual(acl.Read.Groups, []string{"admins", "users", "clients"}) {
		t.Errorf("read.groups = %v", acl.Read.Groups)
	}
}

func TestACLs_SetPermission(t *testing.T) {
	// PUT a single permission. Body shape: {"<perm>": {"actors": [...], "groups": [...]}}.
	srv := cinctest.New(t)
	srv.Handle("PUT /organizations/o/nodes/web01/_acl/update", cinctest.Route{
		Body: `{}`,
		Assert: func(t *testing.T, _ *http.Request, body []byte) {
			var req map[string]ACE
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode: %v (body=%s)", err, body)
			}
			ace, ok := req["update"]
			if !ok {
				t.Fatalf("PUT body missing perm key: %s", body)
			}
			if !reflect.DeepEqual(ace.Groups, []string{"admins", "ops"}) {
				t.Errorf("groups = %v", ace.Groups)
			}
			// Empty actors must serialize as [] not null.
			var raw struct {
				Update struct {
					Actors json.RawMessage `json:"actors"`
				} `json:"update"`
			}
			json.Unmarshal(body, &raw)
			if string(raw.Update.Actors) != "[]" {
				t.Errorf("actors = %s, want []", raw.Update.Actors)
			}
		},
	})
	c := newTestClient(t, srv.Server)
	err := c.ACLs.SetPermission(context.Background(), "nodes", "web01", "update", &ACE{
		Groups: []string{"admins", "ops"},
	})
	if err != nil {
		t.Fatalf("SetPermission: %v", err)
	}
}

func TestACLs_AcceptsAnyObjectType(t *testing.T) {
	// The same endpoint shape is reachable for every object kind.
	srv := cinctest.New(t)
	for _, path := range []string{
		"GET /organizations/o/clients/c1/_acl",
		"GET /organizations/o/groups/devs/_acl",
		"GET /organizations/o/policies/appserver/_acl",
		"GET /organizations/o/containers/nodes/_acl",
	} {
		srv.Handle(path, cinctest.Route{Body: `{}`})
	}
	c := newTestClient(t, srv.Server)
	ctx := context.Background()
	for _, tc := range []struct{ kind, name string }{
		{"clients", "c1"}, {"groups", "devs"},
		{"policies", "appserver"}, {"containers", "nodes"},
	} {
		if _, _, err := c.ACLs.Get(ctx, tc.kind, tc.name); err != nil {
			t.Errorf("Get(%s, %s): %v", tc.kind, tc.name, err)
		}
	}
}

func TestACLs_GetNotFound(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/nodes/missing/_acl",
		cinctest.Route{Status: 404, Body: `{"error":["not found"]}`})
	c := newTestClient(t, srv.Server)
	if _, _, err := c.ACLs.Get(context.Background(), "nodes", "missing"); err == nil {
		t.Fatal("expected 404")
	}
}
