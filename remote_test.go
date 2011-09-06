package selenium

import (
	"strings"
	"testing"
)

var caps = &Capabilities {
	"browserName": "firefox",
}

var executor = ""


func TestStatus(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}
	defer wd.Quit()

	status, err := wd.Status()
	if err != nil {
		t.Fatal(err.String())
	}

	if len(status.OS.Name) == 0 {
		t.Fatal("No OS")
	}
}

func TestNewSession(t *testing.T) {
	wd := &remoteWD{capabilities: caps, executor: DEFAULT_EXECUTOR}
	sid, err := wd.NewSession()
	defer wd.Quit()

	if err != nil {
		t.Fatalf("error in new session - %s", err)
	}

	if len(sid) == 0 {
		t.Fatal("Empty session id")
	}

	if wd.id != sid {
		t.Fatal("Session id mismatch")
	}
}

func TestCurrentWindowHandle(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	handle, err := wd.CurrentWindowHandle()

	if err != nil {
		t.Fatal(err.String())
	}

	if len(handle) == 0 {
		t.Fatal("Empty handle")
	}
}

func TestWindowHandles(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	handles, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Fatal(err.String())
	}

	if len(handles) == 0 {
		t.Fatal("No handles")
	}
}

func TestGet(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	url := "http://www.google.com/"
	err = wd.Get(url)
	if err != nil {
		t.Fatal(err.String())
	}

	newUrl, err := wd.CurrentURL()
	if err != nil {
		t.Fatal(err.String())
	}

	if newUrl != url {
		t.Fatal("%s != %s", newUrl, url)
	}
}

func TestNavigation(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	url1 := "http://www.google.com/"
	err = wd.Get(url1)
	if err != nil {
		t.Fatal(err.String())
	}

	url2 := "http://golang.org/"
	err = wd.Get(url2)
	if err != nil {
		t.Fatal(err.String())
	}

	err = wd.Back()
	if err != nil {
		t.Fatal(err.String())
	}
	url, _ := wd.CurrentURL()
	if url != url1 {
		t.Fatal("back go me to %s (expected %s)", url, url1)
	}
	err = wd.Forward()
	if err != nil {
		t.Fatal(err.String())
	}
	url, _ = wd.CurrentURL()
	if url != url2 {
		t.Fatal("back go me to %s (expected %s)", url, url2)
	}

	err = wd.Refresh()
	if err != nil {
		t.Fatal(err.String())
	}
	url, _ = wd.CurrentURL()
	if url != url2 {
		t.Fatal("back go me to %s (expected %s)", url, url2)
	}
}

func TestTitle(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	_, err = wd.Title()
	if err != nil {
		t.Fatal(err.String())
	}
}

func TestPageSource(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	_, err = wd.PageSource()
	if err != nil {
		t.Fatal(err.String())
	}
}

func TestFindElement(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.google.com")
	elem, err := wd.FindElement(ByName, "btnK")
	if err != nil {
		t.Fatal(err.String())
	}

	we, ok := elem.(*remoteWE)
	if !ok {
		t.Fatal("Can't convert to *remoteWE")
	}

	if len(we.id) == 0 {
		t.Fatal("Empty element")
	}

	if we.parent != wd {
		t.Fatal("Bad parent")
	}
}

func TestFindElements(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.google.com")
	elems, err := wd.FindElements(ByName, "btnK")
	if err != nil {
		t.Fatal(err.String())
	}

	if len(elems) != 1 {
		t.Fatal("Wrong number of elements %d (should be 1)", len(elems))
	}


	we, ok := elems[0].(*remoteWE)
	if !ok {
		t.Fatal("Can't convert to *remoteWE")
	}

	if len(we.id) == 0 {
		t.Fatal("Empty element")
	}

	if we.parent != wd {
		t.Fatal("Bad parent")
	}
}

func TestSendKeys(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	input, err := wd.FindElement(ByName, "p")
	if err != nil {
		t.Fatal(err.String())
	}
	err = input.SendKeys("golang\n")
	if err != nil {
		t.Fatal(err.String())
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err.String())
	}

	if !strings.Contains(source, "The Go Programming Language") {
		t.Fatal("Google can't find Go")
	}

}

func TestClick(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	input, err := wd.FindElement(ByName, "p")
	if err != nil {
		t.Fatal(err.String())
	}
	err = input.SendKeys("golang")
	if err != nil {
		t.Fatal(err.String())
	}

	button, err := wd.FindElement(ById, "search-submit")
	if err != nil {
		t.Fatal(err.String())
	}
	err = button.Click()
	if err != nil {
		t.Fatal(err.String())
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err.String())
	}

	if !strings.Contains(source, "The Go Programming Language") {
		t.Fatal("Google can't find Go")
	}
}

func TestGetCookies(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.google.com")
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err.String())
	}

	if len(cookies) == 0 {
		t.Fatal("No cookies")
	}

	if len(cookies[0].Name) == 0 {
		t.Fatal("Empty cookie")
	}
}

func TestAddCookie(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.google.com")
	cookie := &Cookie{Name: "the nameless cookie", Value: "I have nothing"}
	err = wd.AddCookie(cookie)
	if err != nil {
		t.Fatal(err.String())
	}

	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err.String())
	}
	for _, c := range(cookies) {
		if (c.Name == cookie.Name) && (c.Value == cookie.Value) {
			return
		}
	}

	t.Fatal("Can't find new cookie")
}

func TestDeleteCookie(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.google.com")
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err.String())
	}
	err = wd.DeleteCookie(cookies[0].Name)
	if err != nil {
		t.Fatal(err.String())
	}
	newCookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err.String())
	}
	if len(newCookies) != len(cookies) - 1 {
		t.Fatal("Cookie not deleted")
	}

	for _, c := range(newCookies) {
		if c.Name == cookies[0].Name {
			t.Fatal("Deleted cookie found")
		}
	}

}
func TestLocation(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	button, err := wd.FindElement(ById, "search-submit")
	if err != nil {
		t.Fatal(err.String())
	}

	loc, err := button.Location()
	if err != nil {
		t.Fatal(err.String())
	}

	if (loc.X == 0) || (loc.Y == 0) {
		t.Fatalf("Bad location: %v\n", loc)
	}
}

func TestLocationInView(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	button, err := wd.FindElement(ById, "search-submit")
	if err != nil {
		t.Fatal(err.String())
	}

	loc, err := button.LocationInView()
	if err != nil {
		t.Fatal(err.String())
	}

	if (loc.X == 0) || (loc.Y == 0) {
		t.Fatalf("Bad location: %v\n", loc)
	}
}

func TestSize(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	button, err := wd.FindElement(ById, "search-submit")
	if err != nil {
		t.Fatal(err.String())
	}

	size, err := button.Size()
	if err != nil {
		t.Fatal(err.String())
	}

	if (size.Width == 0) || (size.Height == 0) {
		t.Fatalf("Bad size: %v\n", size)
	}
}

func TestExecuteScript(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	script := "return arguments[0] + arguments[1]"
	args := []interface{}{1, 2}
	reply, err := wd.ExecuteScript(script, args)
	if err != nil {
		t.Fatal(err.String())
	}

	result, ok := reply.(float64)
	if !ok {
		t.Fatal("Not an int reply")
	}

	if result != 3 {
		t.Fatal("Bad result %d (expected 3)", result)
	}
}

func TestScreenshot(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.yahoo.com")
	data, err := wd.Screenshot()
	if err != nil {
		t.Fatal(err.String())
	}

	if len(data) == 0 {
		t.Fatal("Empty reply")
	}
}

func TestIsSelected(t *testing.T) {
	wd, err := NewRemote(caps, executor, "")
	if err != nil {
		t.Fatal("No session")
	}

	defer wd.Quit()

	wd.Get("http://www.google.com/advanced_image_search?hl=en")
	elem, err := wd.FindElement(ById, "cc_com")
	if err != nil {
		t.Fatal("Can't find element")
	}
	selected, err := elem.IsSelected()
	if err != nil {
		t.Fatal("Can't get selection")
	}

	if selected {
		t.Fatal("Already selected")
	}

	err = elem.Click()
	if err != nil {
		t.Fatalf("Can't click")
	}

	selected, err = elem.IsSelected()
	if err != nil {
		t.Fatal("Can't get selection")
	}

	if !selected {
		t.Fatal("Not selected")
	}
}
