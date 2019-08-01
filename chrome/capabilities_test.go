package chrome

import (
	"encoding/json"
	"testing"
)

func TestEmptyCapabilities(t *testing.T) {
	data, err := json.Marshal(Capabilities{})
	if err != nil {
		t.Fatalf("json.Marshal(Capabilities{}) return error: %v", err)
	}
	got, want := string(data), `{"w3c":false}`
	if got != want {
		t.Fatalf("json.Marshal(Capabilities{}) = %q, want %q", got, want)
	}
}
