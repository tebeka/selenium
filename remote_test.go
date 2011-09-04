package selenium

import (
//	"strings"
	"testing"
)

var caps = &Capabilities {
	"browserName": "firefox",
}

/*
func TestStatus(t *testing.T) {
	wd, err := NewRemote(nil, "", nil)
	if err != nil {
		t.Error(err.String())
	}

	status, err := wd.Status()
	if err != nil {
		t.Error(err.String())
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
		t.Error(err.String())
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
		t.Error(err.String())
	}

	if len(handles) == 0 {
		t.Error("No handles")
	}
}

func TestGet(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	url := "http://www.google.com/"
	err := wd.Get(url)
	if err != nil {
		t.Error(err.String())
	}

	newUrl, err := wd.CurrentURL()
	if err != nil {
		t.Error(err.String())
	}

	if newUrl != url {
		t.Error("%s != %s", newUrl, url)
	}
}

func TestNavigation(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	url1 := "http://www.google.com/"
	err := wd.Get(url1)
	if err != nil {
		t.Error(err.String())
	}

	url2 := "http://golang.org/"
	err = wd.Get(url2)
	if err != nil {
		t.Error(err.String())
	}

	err = wd.Back()
	if err != nil {
		t.Error(err.String())
	}
	url, _ := wd.CurrentURL()
	if url != url1 {
		t.Error("back go me to %s (expected %s)", url, url1)
	}
	err = wd.Forward()
	if err != nil {
		t.Error(err.String())
	}
	url, _ = wd.CurrentURL()
	if url != url2 {
		t.Error("back go me to %s (expected %s)", url, url2)
	}

	err = wd.Refresh()
	if err != nil {
		t.Error(err.String())
	}
	url, _ = wd.CurrentURL()
	if url != url2 {
		t.Error("back go me to %s (expected %s)", url, url2)
	}
}

func TestTitle(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	_, err := wd.Title()
	if err != nil {
		t.Error(err.String())
	}
}

func TestPageSource(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	_, err := wd.PageSource()
	if err != nil {
		t.Error(err.String())
	}
}

func TestFindElement(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	wd.Get("http://www.google.com")
	elem, err := wd.FindElement(ByName, "btnK")
	if err != nil {
		t.Error(err.String())
	}

	we, ok := elem.(*remoteWE)
	if !ok {
		t.Error("Can't convert to *remoteWE")
	}

	if len(we.id) == 0 {
		t.Error("Empty element")
	}

	if we.parent != wd {
		t.Error("Bad parent")
	}
}

func TestFindElements(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	wd.Get("http://www.google.com")
	elems, err := wd.FindElements(ByName, "btnK")
	if err != nil {
		t.Error(err.String())
	}

	if len(elems) != 1 {
		t.Error("Wrong number of elements %d (should be 1)", len(elems))
	}


	we, ok := elems[0].(*remoteWE)
	if !ok {
		t.Error("Can't convert to *remoteWE")
	}

	if len(we.id) == 0 {
		t.Error("Empty element")
	}

	if we.parent != wd {
		t.Error("Bad parent")
	}
}

func TestSendKeys(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	input, err := wd.FindElement(ByName, "p")
	if err != nil {
		t.Error(err.String())
	}
	err = input.SendKeys("golang\n")
	if err != nil {
		t.Error(err.String())
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Error(err.String())
	}

	if !strings.Contains(source, "The Go Programming Language") {
		t.Error("Google can't find Go")
	}

}

func TestClick(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	input, err := wd.FindElement(ByName, "p")
	if err != nil {
		t.Error(err.String())
	}
	err = input.SendKeys("golang")
	if err != nil {
		t.Error(err.String())
	}

	button, err := wd.FindElement(ById, "search-submit")
	if err != nil {
		t.Error(err.String())
	}
	err = button.Click()
	if err != nil {
		t.Error(err.String())
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Error(err.String())
	}

	if !strings.Contains(source, "The Go Programming Language") {
		t.Error("Google can't find Go")
	}
}

func TestGetCookies(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	wd.Get("http://www.google.com")
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Error(err.String())
	}

	if len(cookies) == 0 {
		t.Error("No cookies")
	}

	if len(cookies[0].Name) == 0 {
		t.Error("Empty cookie")
	}
}
*/

func TestAddCookie(t *testing.T) {
	wd, _ := NewRemote(caps, "", nil)
	defer wd.Quit()

	wd.Get("http://www.google.com")
	cookie := &Cookie{Name: "the nameless cookie", Value: "I have nothing"}
	err := wd.AddCookie(cookie)
	if err != nil {
		t.Error(err.String())
	}

	cookies, err := wd.GetCookies()
	if err != nil {
		t.Error(err.String())
	}
	for _, c := range(cookies) {
		if (c.Name == cookie.Name) && (c.Value == cookie.Value) {
			return
		}
	}

	t.Error("Can't find new cookie")
}
