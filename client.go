package cinc

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a Chef/CINC Server API client. It is safe for concurrent use.
type Client struct {
	baseURL    *url.URL
	org        string
	clientName string
	key        *rsa.PrivateKey
	httpClient *http.Client
	opts       options
	clock      func() time.Time

	// Services.
	Nodes             *NodesService
	Roles             *RolesService
	Environments      *EnvironmentsService
	Clients           *ClientsService
	DataBags          *DataBagsService
	Search            *SearchService
	Cookbooks         *CookbooksService
	CookbookArtifacts *CookbookArtifactsService
}

// NewClient builds a Client from cfg and optional Options.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	base, err := url.Parse(strings.TrimRight(cfg.ServerURL, "/"))
	if err != nil || base.Host == "" {
		return nil, fmt.Errorf("cinc: invalid ServerURL %q", cfg.ServerURL)
	}
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	hc := o.httpClient
	if o.skipTLSVerify {
		hc = &http.Client{Timeout: hc.Timeout, Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	}
	c := &Client{
		baseURL: base, org: cfg.Org, clientName: cfg.ClientName,
		key: cfg.Key, httpClient: hc, opts: o, clock: time.Now,
	}
	c.Nodes = &NodesService{client: c}
	c.Roles = &RolesService{client: c}
	c.Environments = &EnvironmentsService{client: c}
	c.Clients = &ClientsService{client: c}
	c.DataBags = &DataBagsService{client: c}
	c.Search = &SearchService{client: c}
	c.Cookbooks = &CookbooksService{client: c}
	c.CookbookArtifacts = &CookbookArtifactsService{client: c}
	return c, nil
}

// orgPath prefixes p with /organizations/<org>.
func (c *Client) orgPath(p string) string {
	return "/organizations/" + c.org + "/" + strings.TrimLeft(p, "/")
}

// timestamp returns the current time as an ISO-8601 UTC string.
func (c *Client) timestamp() string {
	return c.clock().UTC().Format("2006-01-02T15:04:05Z")
}
