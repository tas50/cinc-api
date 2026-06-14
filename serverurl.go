package cinc

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseServerURL splits a Chef/CINC Server URL of the form
// https://host[:port]/organizations/<org> into its base server URL
// (scheme://host[:port]) and organization name.
//
// It is the parsing counterpart to NewClient, which takes the base ServerURL
// and Org separately: callers that hold a single combined URL (as Chef's
// credentials files and CHEF_SERVER_URL store it) use this to derive the two.
func ParseServerURL(raw string) (serverURL, org string, err error) {
	u, parseErr := url.Parse(raw)
	if parseErr != nil || u.Scheme == "" || u.Host == "" {
		return "", "", fmt.Errorf("cinc: invalid server URL %q", raw)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "organizations" || parts[1] == "" {
		return "", "", fmt.Errorf("cinc: server URL %q must end with /organizations/<org>", raw)
	}
	return u.Scheme + "://" + u.Host, parts[1], nil
}
