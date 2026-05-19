// Package signing implements the Chef v1.3 (SHA-256) signed-header protocol.
package signing

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"strings"
)

// ServerAPIVersion is the X-Ops-Server-API-Version the client requests.
const ServerAPIVersion = "1"

// Request is the minimal set of fields needed to sign an HTTP request.
type Request struct {
	Method    string // upper-case HTTP method
	Path      string // request path, will be canonicalized
	Body      []byte // request body, may be nil
	UserID    string // Chef client/user name
	Timestamp string // ISO-8601 UTC, e.g. 2024-01-01T00:00:00Z
}

// ContentHash returns base64(sha256(body)).
func ContentHash(body []byte) string {
	sum := sha256.Sum256(body)
	return base64.StdEncoding.EncodeToString(sum[:])
}

var multiSlash = regexp.MustCompile(`/+`)

// CanonicalPath collapses repeated slashes and strips a trailing slash.
func CanonicalPath(p string) string {
	if p == "" {
		return "/"
	}
	p = multiSlash.ReplaceAllString(p, "/")
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// CanonicalRequest builds the v1.3 string-to-sign for r.
func CanonicalRequest(r Request) string {
	return strings.Join([]string{
		"Method:" + r.Method,
		"Path:" + CanonicalPath(r.Path),
		"X-Ops-Content-Hash:" + ContentHash(r.Body),
		"X-Ops-Sign:version=1.3",
		"X-Ops-Timestamp:" + r.Timestamp,
		"X-Ops-UserId:" + r.UserID,
		"X-Ops-Server-API-Version:" + ServerAPIVersion,
	}, "\n")
}
