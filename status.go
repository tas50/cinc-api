package cinc

import "context"

// ServerStatus is the response from GET /_status. The endpoint is typically
// unauthenticated and is used by load balancers as a health probe.
type ServerStatus struct {
	Status    string            `json:"status"`
	Upstreams map[string]string `json:"upstreams,omitempty"`
	Keygen    KeygenStatus      `json:"keygen"`
}

// KeygenStatus reports on the server's pre-generated keypair pool.
type KeygenStatus struct {
	Keys              int     `json:"keys"`
	Max               int     `json:"max"`
	MaxWorkers        int     `json:"max_workers"`
	CurMaxWorkers     int     `json:"cur_max_workers"`
	Inflight          int     `json:"inflight"`
	AvgCreationTimeMS float64 `json:"avg_creation_time_in_ms"`
}

// StatusService accesses the top-level /_status endpoint.
type StatusService struct{ client *Client }

// Get returns the server's health and keygen pool status.
func (s *StatusService) Get(ctx context.Context) (*ServerStatus, *Response, error) {
	st, resp, err := do[ServerStatus](ctx, s.client, "GET", "/_status", nil)
	return ptrOrNil(st, err), resp, err
}
