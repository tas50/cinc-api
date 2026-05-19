package cinc

// Attributes is a free-form Chef attribute tree. It marshals as a plain JSON
// object and preserves any keys it does not interpret.
type Attributes map[string]any

// Dig walks nested maps along path and returns the value at the leaf.
func (a Attributes) Dig(path ...string) (any, bool) {
	if len(path) == 0 {
		return nil, false
	}
	var cur any = map[string]any(a)
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// GetString returns the string at path, or "" if absent or not a string.
func (a Attributes) GetString(path ...string) string {
	if v, ok := a.Dig(path...); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
