// errors_test.go
package cinc

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorResponse_Is(t *testing.T) {
	cases := map[int]error{
		http.StatusNotFound:     ErrNotFound,
		http.StatusConflict:     ErrConflict,
		http.StatusForbidden:    ErrForbidden,
		http.StatusUnauthorized: ErrUnauthorized,
	}
	for code, sentinel := range cases {
		err := newErrorResponse("GET", "/x", code, []byte(`{"error":["boom"]}`))
		if !errors.Is(err, sentinel) {
			t.Errorf("status %d: errors.Is(%v) = false", code, sentinel)
		}
	}
}

func TestErrorResponse_Message(t *testing.T) {
	err := newErrorResponse("PUT", "/nodes/x", 404, []byte(`{"error":["no node"]}`))
	if got := err.Error(); got == "" ||
		!contains(got, "404") || !contains(got, "no node") {
		t.Fatalf("Error() = %q, want code and message", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		len(sub) == 0 || indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
