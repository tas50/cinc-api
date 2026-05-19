// search_test.go
package cinc

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tas50/cinc-api/internal/cinctest"
)

func TestSearch_Query(t *testing.T) {
	srv := cinctest.New(t)
	srv.Handle("GET /organizations/o/search/node",
		cinctest.Route{
			Body: `{"total":1,"start":0,"rows":[{"name":"web01"}]}`,
			Assert: func(t *testing.T, r *http.Request, _ []byte) {
				q := r.URL.Query()
				if q.Get("q") != "chef_environment:prod" {
					t.Errorf("q = %q", q.Get("q"))
				}
				if q.Get("rows") != "5" {
					t.Errorf("rows = %q", q.Get("rows"))
				}
			}})
	c := newTestClient(t, srv.Server)

	res, _, err := c.Search.Query(context.Background(), "node",
		"chef_environment:prod", WithRows(5))
	if err != nil || res.Total != 1 || len(res.Rows) != 1 {
		t.Fatalf("Query: %+v %v", res, err)
	}
	var n Node
	if err := json.Unmarshal(res.Rows[0], &n); err != nil || n.Name != "web01" {
		t.Fatalf("decode row: %+v %v", n, err)
	}
}

func TestSearch_All(t *testing.T) {
	srv := cinctest.New(t)
	page := 0
	srv.Handle("GET /organizations/o/search/node", cinctest.Route{}) // unused
	srv.Server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if page == 0 {
			page++
			w.Write([]byte(`{"total":3,"start":0,"rows":[{"name":"a"},{"name":"b"}]}`))
			return
		}
		w.Write([]byte(`{"total":3,"start":2,"rows":[{"name":"c"}]}`))
	})
	c := newTestClient(t, srv.Server)
	rows, err := c.Search.SearchAll(context.Background(), "node", "*:*", WithRows(2))
	if err != nil || len(rows) != 3 {
		t.Fatalf("SearchAll: %d rows, %v", len(rows), err)
	}
}
