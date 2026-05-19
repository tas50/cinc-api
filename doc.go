// Package cinc is a Go client for the Chef Infra / CINC Server API.
//
// Construct a Client with NewClient, then call methods on its service fields:
//
//	key, _ := cinc.LoadKeyFile("/etc/chef/client.pem")
//	c, _ := cinc.NewClient(cinc.Config{
//		ServerURL:  "https://chef.example.com",
//		Org:        "myorg",
//		ClientName: "node1",
//		Key:        key,
//	})
//	node, _, err := c.Nodes.Get(context.Background(), "web01")
//
// Requests are authenticated with the Chef v1.3 signed-header protocol.
package cinc
