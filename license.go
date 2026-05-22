package cinc

import "context"

// License is the response from GET /license. It reports node-license usage
// for the Chef Server installation.
type License struct {
	LimitExceeded bool   `json:"limit_exceeded"`
	NodeLicense   int    `json:"node_license"`
	NodeCount     int    `json:"node_count"`
	UpgradeURL    string `json:"upgrade_url,omitempty"`
}

// LicenseService accesses the top-level /license endpoint.
type LicenseService struct{ client *Client }

// Get returns the server's node-license usage.
func (s *LicenseService) Get(ctx context.Context) (*License, *Response, error) {
	lic, resp, err := do[License](ctx, s.client, "GET", "/license", nil)
	return ptrOrNil(lic, err), resp, err
}
