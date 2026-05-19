package cinc

import "context"

// Role is a Chef role object.
type Role struct {
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	RunList            []string            `json:"run_list"`
	DefaultAttributes  Attributes          `json:"default_attributes,omitempty"`
	OverrideAttributes Attributes          `json:"override_attributes,omitempty"`
	EnvRunLists        map[string][]string `json:"env_run_lists,omitempty"`
}

// RolesService accesses the /roles endpoints.
type RolesService struct{ client *Client }

func (s *RolesService) res() crud[Role] {
	return crud[Role]{client: s.client, path: "/roles"}
}

// Get retrieves a role by name.
func (s *RolesService) Get(ctx context.Context, name string) (*Role, *Response, error) {
	r, resp, err := s.res().get(ctx, name)
	return ptrOrNil(r, err), resp, err
}

// Create creates a new role.
func (s *RolesService) Create(ctx context.Context, r *Role) (*Role, *Response, error) {
	created, resp, err := s.res().create(ctx, r.Name, r)
	return ptrOrNil(created, err), resp, err
}

// Update replaces an existing role.
func (s *RolesService) Update(ctx context.Context, r *Role) (*Role, *Response, error) {
	updated, resp, err := s.res().update(ctx, r.Name, r)
	return ptrOrNil(updated, err), resp, err
}

// Delete removes a role by name.
func (s *RolesService) Delete(ctx context.Context, name string) (*Response, error) {
	return s.res().remove(ctx, name)
}

// List returns the role name->URL index.
func (s *RolesService) List(ctx context.Context) (map[string]string, *Response, error) {
	return s.res().list(ctx)
}
