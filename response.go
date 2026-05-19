package cinc

import "net/http"

// Response wraps the raw HTTP response returned by an API call.
type Response struct {
	HTTPResponse *http.Response // the underlying response (body already closed)
	StatusCode   int            // convenience copy of HTTPResponse.StatusCode
}
