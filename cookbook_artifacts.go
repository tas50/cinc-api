package cinc

import "context"

// CookbookArtifactVersion is one identifier entry in an artifact list.
type CookbookArtifactVersion struct {
	Identifier string `json:"identifier"`
	URL        string `json:"url"`
}

// CookbookArtifactListEntry is the list response value for one artifact name.
type CookbookArtifactListEntry struct {
	URL      string                    `json:"url"`
	Versions []CookbookArtifactVersion `json:"versions"`
}

// CookbookArtifactsService accesses the /cookbook_artifacts endpoints. These
// are the Policyfile-mode, content-addressed cookbook variant.
type CookbookArtifactsService struct{ client *Client }

// List returns all cookbook artifacts and their identifiers.
func (s *CookbookArtifactsService) List(ctx context.Context) (map[string]CookbookArtifactListEntry, *Response, error) {
	return do[map[string]CookbookArtifactListEntry](ctx, s.client, "GET",
		s.client.orgPath("/cookbook_artifacts"), nil)
}

// Get retrieves a single cookbook artifact by name and identifier.
func (s *CookbookArtifactsService) Get(ctx context.Context, name, identifier string) (*Cookbook, *Response, error) {
	cb, resp, err := do[Cookbook](ctx, s.client, "GET",
		s.client.orgPath("/cookbook_artifacts/"+name+"/"+identifier), nil)
	return ptrOrNil(cb, err), resp, err
}

// Delete removes a single cookbook artifact.
func (s *CookbookArtifactsService) Delete(ctx context.Context, name, identifier string) (*Response, error) {
	_, resp, err := do[map[string]any](ctx, s.client, "DELETE",
		s.client.orgPath("/cookbook_artifacts/"+name+"/"+identifier), nil)
	return resp, err
}
