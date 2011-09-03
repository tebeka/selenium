package selenium

import (
	"testing"
)

var caps = &Capabilities {
	"browserName": "firefox",
}

func TestStatus(t *testing.T) {
	wd, err := NewRemote(nil, "", nil)
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
	wd := &remoteWD{Capabilities: caps, Executor: DEFAULT_EXECUTOR}
	sid, err := wd.NewSession()
	defer wd.Quit()

	if err != nil {
		t.Errorf("error in new session - %s", err)
	}

	if len(sid) == 0 {
		t.Error("Empty session id")
	}

	if wd.SessionId != sid {
		t.Error("Session id mismatch")
	}
}

func TestCurrentWindowHandle(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	handle, err := wd.CurrentWindowHandle()

	if err != nil {
		t.Error(err)
	}

	if len(handle) == 0 {
		t.Error("Empty handle")
	}
}

func TestWindowHandles(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	handles, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Error(err)
	}

	if len(handles) == 0 {
		t.Error("No handles")
	}
}
