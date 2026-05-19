// attributes_test.go
package cinc

import (
	"encoding/json"
	"testing"
)

func TestAttributes_DigAndGet(t *testing.T) {
	a := Attributes{"kernel": map[string]any{"machine": "x86_64", "bits": float64(64)}}
	if v, ok := a.Dig("kernel", "machine"); !ok || v != "x86_64" {
		t.Fatalf("Dig = %v %v", v, ok)
	}
	if got := a.GetString("kernel", "machine"); got != "x86_64" {
		t.Fatalf("GetString = %q", got)
	}
	if _, ok := a.Dig("kernel", "missing"); ok {
		t.Fatal("expected miss")
	}
}

func TestAttributes_RoundTripPreservesUnknown(t *testing.T) {
	src := []byte(`{"a":1,"nested":{"keep":true},"x":["y"]}`)
	var a Attributes
	if err := json.Unmarshal(src, &a); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var got, want map[string]any
	json.Unmarshal(out, &got)
	json.Unmarshal(src, &want)
	if len(got) != len(want) || got["a"] != want["a"] {
		t.Fatalf("round trip lost data: %s", out)
	}
}
