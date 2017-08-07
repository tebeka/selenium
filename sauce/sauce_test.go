package sauce

import (
	"encoding/json"
	"testing"
)

func TestEmptyCapabilities(t *testing.T) {
	buf, err := json.Marshal(&Capabilities{})
	if err != nil {
		t.Fatalf("json.Marshal(&Capabilities{}) returned error: %s", err)
	}
	if string(buf) != "{}" {
		t.Fatalf("json.Marshal(&Capabilities{}) returned %q, expected '{}'", buf)
	}
}
