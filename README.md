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

| Service                | Path                                          | Methods                                                 |
| ---------------------- | --------------------------------------------- | ------------------------------------------------------- |
| `c.Nodes`              | `/nodes`                                      | List / Get / Create / Update / Delete                   |
| `c.Roles`              | `/roles`                                      | List / Get / Create / Update / Delete                   |
| `c.Environments`      | `/environments`                               | List / Get / Create / Update / Delete                   |
| `c.Clients`            | `/clients`                                    | List / Get / Create / Update / Delete                   |
| `c.DataBags`          | `/data`                                       | List / Create / Delete; per-bag Items handle for CRUD   |
| `c.Search`             | `/search/INDEX`                              | `Query` (with `WithStart`/`WithRows`/`WithPartial`), `SearchAll` |
| `c.Cookbooks`          | `/cookbooks`                                  | List / Get / Delete / Upload (sandbox flow) / Download  |
| `c.CookbookArtifacts` | `/cookbook_artifacts`                         | List / Get / Delete / Upload                            |
| `c.Keys`               | `/users/U/keys`, `/clients/C/keys`         | `User(name)` / `Client(name)` → List / Get / Create / Update / Delete |
| `c.Groups`             | `/groups`                                     | List / Get / Create / Update / Delete                   |
| `c.Policies`           | `/policies`                                   | List / Get / Delete / GetRevision / CreateRevision / DeleteRevision |
| `c.PolicyGroups`      | `/policy_groups`                              | List / Get / Delete / GetPolicy / PutPolicy / DeletePolicy |
| `c.Status`             | `/_status`                                    | Get (server health + keygen pool)                       |
| `c.License`            | `/license`                                    | Get (node-license usage)                                |
| `c.Orgs`               | `/organizations` (top-level)                  | List / Get / Create / Update / Delete                   |

Configurable via options: `WithHTTPClient`, `WithUserAgent`,
`WithChefVersion`, `WithSkipTLSVerify`, `WithMaxRetries`. Idempotent GETs
are retried on 5xx and network errors.

## License

See LICENSE.
