package cinc

import "context"

// Environment is a Chef environment object.
type Environment struct {
	Name               string            `json:"name"`
	Description        string            `json:"description,omitempty"`
	CookbookVersions   map[string]string `json:"cookbook_versions,omitempty"`
	DefaultAttributes  Attributes        `json:"default_attributes,omitempty"`
	OverrideAttributes Attributes        `json:"override_attributes,omitempty"`
}

// EnvironmentsService accesses the /environments endpoints.
type EnvironmentsService struct{ client *Client }

func (s *EnvironmentsService) res() crud[Environment] {
	return crud[Environment]{client: s.client, path: "/environments"}
}

// Get retrieves an environment by name.
func (s *EnvironmentsService) Get(ctx context.Context, name string) (*Environment, *Response, error) {
	e, resp, err := s.res().get(ctx, name)
	return ptrOrNil(e, err), resp, err
}

// Create creates a new environment.
func (s *EnvironmentsService) Create(ctx context.Context, e *Environment) (*Environment, *Response, error) {
	created, resp, err := s.res().create(ctx, e.Name, e)
	return ptrOrNil(created, err), resp, err
}

// Update replaces an existing environment.
func (s *EnvironmentsService) Update(ctx context.Context, e *Environment) (*Environment, *Response, error) {
	updated, resp, err := s.res().update(ctx, e.Name, e)
	return ptrOrNil(updated, err), resp, err
}

// Delete removes an environment by name.
func (s *EnvironmentsService) Delete(ctx context.Context, name string) (*Response, error) {
	return s.res().remove(ctx, name)
}

// List returns the environment name->URL index.
func (s *EnvironmentsService) List(ctx context.Context) (map[string]string, *Response, error) {
	return s.res().list(ctx)
}
