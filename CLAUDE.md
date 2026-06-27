# Notes for Claude Code

This is a Go client for the Chef Infra / CINC Server API. The package is
small, flat, and idiomatic — keep it that way.

## Project layout

- Public types and services live at the repo root, one file per service:
  `nodes.go` + `nodes_test.go`, `policies.go` + `policies_test.go`, etc.
- `internal/signing` implements the Chef v1.3 SHA-256 signed-header protocol.
- `internal/cinctest` is the test-only fake-server harness.
- `testdata/test_key.pem` is the shared fixture RSA key used by all tests.

## Adding a new service

1. Create `foo.go` with:
   - The wire types (struct fields use JSON tags; prefer `omitempty`).
   - A `FooService struct{ client *Client }` and its methods.
   - Standard signatures: reads return `(value, *Response, error)`;
     deletes return `(*Response, error)`. Use `ptrOrNil(v, err)` to
     convert a value to a `*T` that is `nil` when `err != nil`.
   - Use the generic helper `do[T](ctx, c, method, path, body)` for
     every request. It signs, retries idempotent calls, decodes JSON,
     and maps non-2xx responses to `*ErrorResponse`.
   - If the resource has plain Get/Create/Update/Delete/List, the
     `crud[T]` helper in `crud.go` saves boilerplate (see `nodes.go`).
2. Wire the service onto `*Client` in `client.go` — declare the field on
   the struct and assign it in `NewClient`. Add it to the
   `TestNewClient_WiresAllServices` check.
3. Use `c.orgPath(p)` for org-scoped paths (`/organizations/<org>/...`).
   Top-level endpoints (e.g. `/_status`, `/license`, `/users`) take the
   absolute path directly — do **not** call `orgPath`.
4. Update the `README.md` Status table in the same PR (see below).

## Testing

- Tests use `internal/cinctest`:
  ```go
  srv := cinctest.New(t)
  srv.Handle("GET /organizations/o/foo", cinctest.Route{Body: `{...}`})
  c := newTestClient(t, srv.Server)
  ```
  `Route.Assert` is a callback for inspecting the incoming request.
  The harness fails any test that issues an unsigned request, so
  signing is exercised on every call automatically.
- For requests outside the harness (path traversal, raw bookshelf
  uploads, etc.) set `srv.Server.Config.Handler` directly.
- Write tests **first**: red → implement → green. New files should land
  with 100% line coverage; the package as a whole sits at ~96% and
  should stay there or higher.
- Run `go test ./... -race -count=2` before committing. For coverage
  including the test-only `cinctest` package use
  `go test ./... -coverpkg=./... -coverprofile=...`.
- `go vet ./...` must be clean.
- **Integration tests live in `integration/`, a *separate* Go module**
  (its own `go.mod`, so the cinc-zero test dependency never reaches
  consumers of this package — the root import stays zero-dependency).
  The root `go test ./...` does **not** run them. Run them with
  `cd integration && go test ./...` (~1s); they boot an in-memory
  cinc-zero server and exercise the real wire protocol end-to-end,
  unlike the `cinctest` fake the unit tests use. Run them when you
  touch the transport, signing, or cookbook-upload paths.

## Auth and transport gotchas

- The client signs every request with the v1.3 SHA-256 header protocol.
  The signed canonical path is stripped of any `?query` — that fix is
  in `transport.doOnce`; do not re-introduce the query string.
- **Pre-signed URLs (cookbook bookshelf, sandbox uploads) must NOT carry
  Chef signing headers.** `c.uploadFile` and `c.downloadFile` use raw
  `httpClient.Do` to enforce this. Any test that hits a bookshelf URL
  must assert `X-Ops-Authorization-1` is absent.
- Retries: GETs are retried on 5xx and on network errors, up to
  `WithMaxRetries(n)` (default 2). Non-GET requests are never retried.
  Context cancellation/deadline never retries.

## Encoding edge cases worth remembering

- `Group.Update` rewraps `Users/Clients/Groups` into the server's
  required `actors: {users, clients, groups}` shape. Nil slices must
  serialize as `[]`, not `null` — `groups.go:nonNil` exists for this.
- `omitempty` does **not** elide nested struct values in Go's
  `encoding/json`. Use `omitzero` (Go 1.24+) or drop the tag when the
  field is a value-type struct.
- Cookbook upload is three-step: `POST /sandboxes` → file PUTs to the
  pre-signed URLs the server returned → manifest PUT. `uploadCookbook`
  in `cookbooks.go` is shared with cookbook artifacts.
- Search supports `WithStart`, `WithRows`, `WithPartial`. Passing
  `WithPartial` switches the underlying request from GET to POST with
  the projection map as the body — this is server-required, not a
  client choice.

## PR conventions

- Whenever you open a PR that adds, removes, or renames a public
  surface (service, exported method, option, type), update `README.md`
  in the same PR so the **Status** table matches what the package
  actually exposes.
- Keep commit messages explanatory and terse — see `git log` for the
  established style.
- The PR description should include a **Test plan** checklist.
- One feature per PR. Multiple endpoint families should ship as
  separate PRs unless they are tightly coupled (e.g. Keys + Groups
  shipped together because they share ACL semantics).
