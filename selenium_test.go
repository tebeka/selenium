package selenium

import (
	"testing"
)

func TestStatus(t *testing.T) {
	wd, err := New(nil, "", nil)
	if err != nil {
		t.Error(err)
	}

	status, err := wd.Status()
	if err != nil {
		t.Error(err)
	}

	if len(status.OS.Name) == 0 {
		t.Error("No OS")
	}
}

func TestNewSession(t *testing.T) {
	caps, _ := NewCapabilities("browserName", "firefox")

	wd, _ := New(caps, "", nil)
	wd.NewSession()
}


func TestURL(t *testing.T) {
	params := Params{SessionId: "17"}
	url, err := CMD_QUIT.URL(&params)

	if err != nil {
		t.Error(err)
	}

	if url != "/session/17" {
		t.Errorf("Bad url: %s", url)
	}
}
