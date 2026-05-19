package cinc

import "context"

// APIClient is a Chef API client (node or validator identity). It is named
// APIClient to avoid colliding with the package's own Client type.
type APIClient struct {
	Name      string  `json:"name"`
	Validator bool    `json:"validator"`
	ChefKey   ChefKey `json:"chef_key"`
}

// ChefKey holds key material returned when a client is created.
type ChefKey struct {
	Name       string `json:"name,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	ExpiresAt  string `json:"expiration_date,omitempty"`
}

// ClientsService accesses the /clients endpoints.
type ClientsService struct{ client *Client }

func (s *ClientsService) res() crud[APIClient] {
	return crud[APIClient]{client: s.client, path: "/clients"}
}

// Get retrieves a client by name.
func (s *ClientsService) Get(ctx context.Context, name string) (*APIClient, *Response, error) {
	cl, resp, err := s.res().get(ctx, name)
	return ptrOrNil(cl, err), resp, err
}

// Create creates a new client and returns its generated key material.
func (s *ClientsService) Create(ctx context.Context, cl *APIClient) (*APIClient, *Response, error) {
	created, resp, err := s.res().create(ctx, cl.Name, cl)
	return ptrOrNil(created, err), resp, err
}

// Update replaces an existing client.
func (s *ClientsService) Update(ctx context.Context, cl *APIClient) (*APIClient, *Response, error) {
	updated, resp, err := s.res().update(ctx, cl.Name, cl)
	return ptrOrNil(updated, err), resp, err
}

// Delete removes a client by name.
func (s *ClientsService) Delete(ctx context.Context, name string) (*Response, error) {
	return s.res().remove(ctx, name)
}

// List returns the client name->URL index.
func (s *ClientsService) List(ctx context.Context) (map[string]string, *Response, error) {
	return s.res().list(ctx)
}
