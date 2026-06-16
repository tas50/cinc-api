# cinc-api

A modern, idiomatic Go client for the Chef Infra / CINC Server API.

## Install

    go get github.com/tas50/cinc-api

## Usage

    key, _ := cinc.LoadKeyFile("/etc/chef/client.pem")
    c, err := cinc.NewClient(cinc.Config{
        ServerURL:  "https://chef.example.com",
        Org:        "myorg",
        ClientName: "node1",
        Key:        key,
    })
    node, _, err := c.Nodes.Get(context.Background(), "web01")

## Status

Authentication uses the Chef v1.3 SHA-256 signed-header protocol. The
following endpoint families are implemented:

| Service              | Path                                  | Methods                                                              |
| -------------------- | ------------------------------------- | -------------------------------------------------------------------- |
| `c.ACLs`             | `/<object>/<name>/_acl`               | Get / SetPermission                                                  |
| `c.Clients`          | `/clients`                            | List / Get / Create / Update / Delete / Reregister                   |
| `c.Containers`       | `/containers`                         | List / Get / Create / Delete                                         |
| `c.Cookbooks`        | `/cookbooks`                          | List / Get (with metadata) / Delete / Upload (sandbox flow) / Download |
| `c.CookbookArtifacts`| `/cookbook_artifacts`                 | List / Get (with metadata) / Delete / Upload                         |
| `c.DataBags`         | `/data`                               | List / Create / Delete; per-bag Items handle for CRUD                |
| `c.Environments`     | `/environments`                       | List / Get / Create / Update / Delete                                |
| `c.Groups`           | `/groups`                             | List / Get / Create / Update / Delete                                |
| `c.Keys`             | `/users/U/keys`, `/clients/C/keys`    | `User(name)` / `Client(name)` → List / Get / Create / Update / Delete |
| `c.License`          | `/license`                            | Get (node-license usage)                                             |
| `c.Nodes`            | `/nodes`                              | List / Get / Create / Update / Delete                                |
| `c.Orgs`             | `/organizations` (top-level)          | List / Get / Create / Update / Delete                                |
| `c.Policies`         | `/policies`                           | List / Get / Delete / GetRevision / CreateRevision / DeleteRevision / PushRevision |
| `c.PolicyGroups`     | `/policy_groups`                      | List / Get / Delete / GetPolicy / PutPolicy / DeletePolicy           |
| `c.RequiredRecipe`   | `/required_recipe`                    | Get (returns Ruby text/plain)                                        |
| `c.Roles`            | `/roles`                              | List / Get / Create / Update / Delete                                |
| `c.Search`           | `/search/INDEX`                       | `Query` (with `WithStart`/`WithRows`/`WithPartial`), `SearchAll`     |
| `c.Status`           | `/_status`                            | Get (server health + keygen pool)                                    |
| `c.Users`            | `/users` (top-level)                  | List / Get / Create / Update / Delete                                |

Configurable via options: `WithHTTPClient`, `WithUserAgent`,
`WithChefVersion`, `WithSkipTLSVerify`, `WithMaxRetries`. Idempotent GETs
are retried on 5xx and network errors.

### Helpers

Standalone helpers for working with Chef/CINC identities and the node object
model, so callers don't re-encode server conventions:

- `ParseServerURL(raw)` — split `https://host/organizations/<org>` into the
  base server URL and org (the inverse of `NewClient`'s `ServerURL`/`Org`).
- `GenerateKeyPair()` — mint a 2048-bit RSA key pair as PEM (the generation
  counterpart to `ParseKey`/`LoadKeyFile`).
- `Node` accessors — `Tags`/`SetTags`/`AddTags`/`RemoveTags` (stored at
  `normal.tags`), `AddRunListItems`/`RemoveRunListItems`, and
  `Attribute`/`AttributeString` (precedence-aware lookup, dotted paths).
- `Clients.Reregister(name)` — regenerate a client's `default` key and return
  the new private key.
- `ParsePolicyfileLock(data)` / `LoadPolicyfileLock(path)` — parse a
  `Policyfile.lock.json` into a `PolicyRevision`.
- `Policies.PushRevision(lockJSON, group, cookbooks)` — the server-side half of
  `chef push`: upload each pinned cookbook as an artifact, then associate the
  revision with a policy group. The lock bytes are sent verbatim so no fields
  are lost.

## License

See LICENSE.
