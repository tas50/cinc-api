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

v0.1 covers nodes, roles, environments, clients, data bags, search, and
cookbooks (including cookbook artifacts). Authentication uses the Chef v1.3
SHA-256 signed-header protocol.

## License

See LICENSE.
