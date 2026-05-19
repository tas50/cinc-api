package cinc

import "context"

// Node is a Chef node object.
type Node struct {
	Name        string     `json:"name"`
	Environment string     `json:"chef_environment,omitempty"`
	RunList     []string   `json:"run_list"`
	Normal      Attributes `json:"normal,omitempty"`
	Default     Attributes `json:"default,omitempty"`
	Override    Attributes `json:"override,omitempty"`
	Automatic   Attributes `json:"automatic,omitempty"`
	PolicyName  string     `json:"policy_name,omitempty"`
	PolicyGroup string     `json:"policy_group,omitempty"`
}

// NodesService accesses the /nodes endpoints.
type NodesService struct{ client *Client }

func (s *NodesService) res() crud[Node] {
	return crud[Node]{client: s.client, path: "/nodes"}
}

// Get retrieves a node by name.
func (s *NodesService) Get(ctx context.Context, name string) (*Node, *Response, error) {
	n, resp, err := s.res().get(ctx, name)
	return ptrOrNil(n, err), resp, err
}

// Create creates a new node.
func (s *NodesService) Create(ctx context.Context, n *Node) (*Node, *Response, error) {
	created, resp, err := s.res().create(ctx, n.Name, n)
	return ptrOrNil(created, err), resp, err
}

// Update replaces an existing node.
func (s *NodesService) Update(ctx context.Context, n *Node) (*Node, *Response, error) {
	updated, resp, err := s.res().update(ctx, n.Name, n)
	return ptrOrNil(updated, err), resp, err
}

// Delete removes a node by name.
func (s *NodesService) Delete(ctx context.Context, name string) (*Response, error) {
	return s.res().remove(ctx, name)
}

// List returns the node name->URL index.
func (s *NodesService) List(ctx context.Context) (map[string]string, *Response, error) {
	return s.res().list(ctx)
}

// ptrOrNil returns &v on success, nil if err != nil.
func ptrOrNil[T any](v T, err error) *T {
	if err != nil {
		return nil
	}
	return &v
}
