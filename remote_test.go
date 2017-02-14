package selenium

import (
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
)

var (
	selenium2Path          = flag.String("selenium2_path", "vendor/selenium-server-standalone-2.53.1.jar", "The path to the Selenium 2 server JAR. If empty or the file is not present, Firefox tests on Selenium 2 will not be run.")
	firefoxBinarySelenium2 = flag.String("firefox_binary_for_selenium2", "vendor/firefox-47/firefox", "The name of the Firefox binary for Selenium 2 tests or the path to it. If the name does not contain directory separators, the PATH will be searched.")

	chromeDriverPath = flag.String("chrome_driver_path", "vendor/chromedriver-linux64-2.27", "The path to the ChromeDriver binary. If empty of the file is not present, Chrome tests will not be run.")
	chromeBinary     = flag.String("chrome_binary", "chromium", "The name of the Chrome binary or the path to it. If name is not an exact path, the PATH will be searched.")

	useDocker          = flag.Bool("docker", false, "If set, run the tests in a Docker container.")
	runningUnderDocker = flag.Bool("running_under_docker", false, "This is set by the Docker test harness and should not be needed otherwise.")

	startFrameBuffer = flag.Bool("start_frame_buffer", true, "If true, start an Xvfb subprocess and run the browsers in that X server.")

	serverURL string
)

func TestMain(m *testing.M) {
	flag.Parse()
	s := httptest.NewServer(http.HandlerFunc(handler))
	serverURL = s.URL
	defer s.Close()
	os.Exit(m.Run())
}

func pickUnusedPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		return 0, err
	}
	return port, nil
}

type config struct {
	addr, browser, path string
	seleniumVersion     semver.Version
}

func TestChrome(t *testing.T) {
	if *useDocker {
		t.Skip("Skipping Chrome tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(*chromeBinary); err != nil {
		path, err := exec.LookPath(*chromeBinary)
		if err != nil {
			t.Skipf("Skipping Chrome tests because binary %q not found", *chromeBinary)
		}
		*chromeBinary = path
	}
	if _, err := os.Stat(*chromeDriverPath); err != nil {
		t.Skipf("Skipping Chrome tests because ChromeDriver not found at path %q", *chromeDriverPath)
	}

	var opts []ServiceOption
	if *startFrameBuffer {
		opts = append(opts, StartFrameBuffer())
	}
	if testing.Verbose() {
		SetDebug(true)
		opts = append(opts, Output(os.Stderr))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}

	s, err := NewChromeDriverService(*chromeDriverPath, port, opts...)
	if err != nil {
		t.Fatalf("Error starting the ChromeDriver server: %v", err)
	}
	c := config{
		addr:    fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port),
		browser: "chrome",
		path:    *chromeBinary,
	}
	runTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the ChromeDriver service: %v", err)
	}
}

func TestFirefoxSelenium2(t *testing.T) {
	if *useDocker {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(*selenium2Path); err != nil {
		t.Skipf("Skipping Firefox tests using Selenium 2 because Selenium WebDriver JAR not found at path %q", *selenium2Path)
	}
	runFirefoxTests(t, *selenium2Path, config{
		seleniumVersion: semver.MustParse("2.0.0"),
		path:            *firefoxBinarySelenium2,
	})
}

func runFirefoxTests(t *testing.T, seleniumPath string, c config) {
	c.browser = "firefox"

	if s, err := os.Stat(c.path); err != nil || !s.Mode().IsRegular() {
		if path, err := exec.LookPath(c.path); err == nil {
			c.path = path
		} else {
			t.Skipf("Skipping Firefox tests because binary %q not found", c.path)
		}
	}
	var opts []ServiceOption
	if *startFrameBuffer {
		opts = append(opts, StartFrameBuffer())
	}
	if testing.Verbose() {
		SetDebug(true)
		opts = append(opts, Output(os.Stderr))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}

	s, err := NewSeleniumService(*selenium2Path, port, opts...)
	if err != nil {
		t.Fatalf("Error starting the Selenium server: %v", err)
	}
	c.addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)

	runTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the Selenium service: %v", err)
	}
}

func TestDocker(t *testing.T) {
	if *runningUnderDocker {
		return
	}
	if !*useDocker {
		t.Skip("Skipping Docker tests because --docker was not specified.")
	}

	args := []string{"build", "-t", "go-selenium", "testing/"}
	if out, err := exec.Command("docker", args...).CombinedOutput(); err != nil {
		t.Logf("Output from `docker %s`:\n%s", strings.Join(args, " "), string(out))
		t.Fatalf("Building Docker container failed: %v", err)
	}

	pathToMount := os.Getenv("GOPATH")
	if strings.Contains(pathToMount, ":") {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("os.Getwd() returned error: %v", err)
		}
		pathToMount = filepath.Join(cwd, "../../../..")
	}
	// TODO(minusnine): pass through relevant flags to docker-test.sh to be
	// passed to go test.
	cmd := exec.Command("docker", "run", fmt.Sprintf("--volume=%s:/code", pathToMount), "--workdir=/code/src/github.com/tebeka/selenium", "go-selenium", "testing/docker-test.sh")
	if testing.Verbose() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker run failed: %v", err)
	}
}

func newTestCapabilities(c config) Capabilities {
	caps := Capabilities{
		"browserName": c.browser,
	}
	switch c.browser {
	case "chrome":
		chrCaps := chrome.Capabilities{
			Path: c.path,
			Args: []string{
				// This flag is needed to test against Chrome binaries that are not the
				// default installation. The sandbox requires a setuid binary.
				"--no-sandbox",
				// This flag is needed for Chrome versions > 56. However, this flag is
				// deprecated in Chrome 57+. Therefore, this API currently cannot
				// support Chrome 57+.
				//
				// https://bugs.chromium.org/p/chromedriver/issues/detail?id=1625
				//
				// TODO(minusnine): Standardize on the Chrome version to use for
				// testing.
				"--disable-extensions",
			},
		}
		caps.AddChrome(chrCaps)
	case "firefox":
		f := firefox.Capabilities{}
		if c.path != "" {
			// Selenium 2 uses this option to specify the path to the Firefox binary.
			caps["firefox_binary"] = c.path
			f.Binary = c.path
		}
		if testing.Verbose() {
			f.Log = &firefox.Log{
				Level: firefox.Trace,
			}
		}
		caps.AddFirefox(f)
	}
	return caps
}

func newRemote(t *testing.T, c config) WebDriver {
	caps := newTestCapabilities(c)
	wd, err := NewRemote(caps, c.addr)
	if err != nil {
		t.Fatalf("NewRemote(%+v, %q) returned error: %v", caps, c.addr, err)
	}
	return wd
}

func quitRemote(t *testing.T, wd WebDriver) {
	if err := wd.Quit(); err != nil {
		t.Errorf("wd.Quit() returned error: %v", err)
	}
}

func runTest(f func(*testing.T, config), c config) func(*testing.T) {
	return func(t *testing.T) {
		f(t, c)
	}
}

func runTests(t *testing.T, c config) {
	t.Run("Status", runTest(testStatus, c))
	t.Run("NewSession", runTest(testNewSession, c))
	t.Run("ExtendedErrorMessage", runTest(testExtendedErrorMessage, c))
	t.Run("Capabilities", runTest(testCapabilities, c))
	t.Run("SetAsyncScriptTimeout", runTest(testSetAsyncScriptTimeout, c))
	t.Run("SetImplicitWaitTimeout", runTest(testSetImplicitWaitTimeout, c))
	t.Run("SetPageLoadTimeout", runTest(testSetPageLoadTimeout, c))
	t.Run("CurrentWindowHandle", runTest(testCurrentWindowHandle, c))
	t.Run("WindowHandles", runTest(testWindowHandles, c))
	t.Run("Get", runTest(testGet, c))
	t.Run("Navigation", runTest(testNavigation, c))
	t.Run("Title", runTest(testTitle, c))
	t.Run("PageSource", runTest(testPageSource, c))
	t.Run("FindElement", runTest(testFindElement, c))
	t.Run("FindElements", runTest(testFindElements, c))
	t.Run("SendKeys", runTest(testSendKeys, c))
	t.Run("Click", runTest(testClick, c))
	t.Run("GetCookies", runTest(testGetCookies, c))
	t.Run("AddCookie", runTest(testAddCookie, c))
	t.Run("DeleteCookie", runTest(testDeleteCookie, c))
	t.Run("Location", runTest(testLocation, c))
	t.Run("LocationInView", runTest(testLocationInView, c))
	t.Run("Size", runTest(testSize, c))
	t.Run("ExecuteScript", runTest(testExecuteScript, c))
	t.Run("ExecuteScriptOnElement", runTest(testExecuteScriptOnElement, c))
	t.Run("ExecuteScriptWithNilArgs", runTest(testExecuteScriptWithNilArgs, c))
	t.Run("Screenshot", runTest(testScreenshot, c))
	t.Run("Log", runTest(testLog, c))
	t.Run("IsSelected", runTest(testIsSelected, c))
	t.Run("IsDisplayed", runTest(testIsDisplayed, c))
	t.Run("GetAttributeNotFound", runTest(testGetAttributeNotFound, c))
	t.Run("MaximizeWindow", runTest(testMaximizeWindow, c))
	t.Run("ResizeWindow", runTest(testResizeWindow, c))
	t.Run("KeyDownUp", runTest(testKeyDownUp, c))
	t.Run("CSSProperty", runTest(testCSSProperty, c))
}

func testStatus(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	status, err := wd.Status()
	if err != nil {
		t.Fatal(err)
	}

	if len(status.OS.Name) == 0 {
		t.Fatal("No OS")
	}
}

func testNewSession(t *testing.T, c config) {
	// Bypass NewRemote which itself calls NewSession internally.
	wd := &remoteWD{
		capabilities: newTestCapabilities(c),
		urlPrefix:    c.addr,
	}
	defer func() {
		if err := wd.Quit(); err != nil {
			t.Errorf("wd.Quit() returned error: %v", err)
		}
	}()

	sid, err := wd.NewSession()
	if err != nil {
		t.Fatalf("error in new session - %s", err)
	}
	defer func() {
		if err := wd.Close(); err != nil {
			t.Errorf("wd.Close() returned error: %v", err)
		}
	}()

	if len(sid) == 0 {
		t.Fatal("Empty session id")
	}

	if wd.id != sid {
		t.Fatal("Session id mismatch")
	}

	if wd.SessionID() != sid {
		t.Fatalf("Got session id mismatch %s != %s", sid, wd.SessionID())
	}
}

func testExtendedErrorMessage(t *testing.T, c config) {
	// Bypass NewRemote which itself calls NewSession internally.
	wd := &remoteWD{
		capabilities: newTestCapabilities(c),
		urlPrefix:    c.addr,
	}
	err := wd.Close()
	if err == nil {
		t.Error("wd.Close() returned nil, expected error")
	}
	switch c.browser {
	case "firefox":
		want := "unknown error:"
		if !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("Got error %q, expected error to start with %q", err, want)
		}
	case "chrome":
		want := "unknown error - 6: no such session"
		if !strings.HasPrefix(err.Error(), want) {
			t.Fatalf("Got error %q, expected error to start with %q", err, want)
		}
	}
}

func testCapabilities(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	caps, err := wd.Capabilities()
	if err != nil {
		t.Fatal(err)
	}

	if caps["browserName"] != c.browser {
		t.Fatalf("bad browser name - %s (should be %s)", caps["browserName"], c.browser)
	}
}

func testSetAsyncScriptTimeout(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.SetAsyncScriptTimeout(200); err != nil {
		t.Fatal(err)
	}
}

func testSetImplicitWaitTimeout(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.SetImplicitWaitTimeout(200); err != nil {
		t.Fatal(err)
	}
}

func testSetPageLoadTimeout(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.SetPageLoadTimeout(200); err != nil {
		t.Fatal(err)
	}
}

func testCurrentWindowHandle(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	handle, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Fatal(err)
	}

	if len(handle) == 0 {
		t.Fatal("Empty handle")
	}
}

func testWindowHandles(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	handles, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Fatal(err)
	}

	if len(handles) == 0 {
		t.Fatal("No handles")
	}
}

func testGet(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatal(err)
	}

	newURL, err := wd.CurrentURL()
	if err != nil {
		t.Fatal(err)
	}

	if newURL != serverURL+"/" {
		t.Fatalf("%s != %s", newURL, serverURL)
	}
}

func testNavigation(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	url1 := serverURL
	if err := wd.Get(url1); err != nil {
		t.Fatal(err)
	}

	url2 := serverURL + "/other"
	if err := wd.Get(url2); err != nil {
		t.Fatal(err)
	}

	if err := wd.Back(); err != nil {
		t.Fatal(err)
	}
	url, err := wd.CurrentURL()
	if err != nil {
		t.Fatalf("wd.CurrentURL() returned error: %v", err)
	}
	if url != url1+"/" {
		t.Fatalf("back got me to %s (expected %s/)", url, url1)
	}
	if err := wd.Forward(); err != nil {
		t.Fatal(err)
	}
	url, err = wd.CurrentURL()
	if err != nil {
		t.Fatalf("wd.CurrentURL() returned error: %v", err)
	}
	if url != url2 {
		t.Fatalf("forward got me to %s (expected %s)", url, url2)
	}

	if err := wd.Refresh(); err != nil {
		t.Fatal(err)
	}
	url, err = wd.CurrentURL()
	if err != nil {
		t.Fatalf("wd.CurrentURL() returned error: %v", err)
	}
	if url != url2 {
		t.Fatalf("refresh got me to %s (expected %s)", url, url2)
	}
}

func testTitle(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	title, err := wd.Title()
	if err != nil {
		t.Fatal(err)
	}

	expectedTitle := "Go Selenium Test Suite"
	if title != expectedTitle {
		t.Fatalf("Bad title %s, should be %s", title, expectedTitle)
	}
}

func testPageSource(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatal(err)
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, "The home page.") {
		t.Fatalf("Bad source\n%s", source)
	}
}

func testFindElement(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	elem, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
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

func testFindElements(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	elems, err := wd.FindElements(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}

	if len(elems) != 1 {
		t.Fatalf("Wrong number of elements %d (should be 1)", len(elems))
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

func testSendKeys(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	input, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}
	const query = "golang"
	if err := input.SendKeys(query + EnterKey); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	source, err := wd.PageSource()
	if err != nil {
		t.Fatalf("wd.PageSource() returned error: %v", err)
	}

	if !strings.Contains(source, searchContents) {
		t.Fatalf("Can't find %q on page after searching for %q", searchContents, query)
	}

	if !strings.Contains(source, query) {
		t.Fatalf("Can't find search query %q in source", query)
	}
}

func testClick(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	input, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}
	const query = "golang"
	if err = input.SendKeys(query); err != nil {
		t.Fatal(err)
	}

	button, err := wd.FindElement(ByID, "submit")
	if err != nil {
		t.Fatal(err)
	}
	if err = button.Click(); err != nil {
		t.Fatal(err)
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, searchContents) {
		t.Fatalf("Can't find %q on page after searching for %q", searchContents, query)
	}
}

func testGetCookies(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) == 0 {
		t.Fatal("No cookies")
	}

	if len(cookies[0].Name) == 0 {
		t.Fatal("Empty cookie")
	}
}

func testAddCookie(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	cookie := &Cookie{
		Name:   "the nameless cookie",
		Value:  "I have nothing",
		Expiry: math.MaxUint32,
	}
	if err := wd.AddCookie(cookie); err != nil {
		t.Fatal(err)
	}

	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cookies {
		if (c.Name == cookie.Name) && (c.Value == cookie.Value) && (c.Expiry == math.MaxUint32) {
			return
		}
	}

	t.Fatal("Can't find new cookie")
}

func testDeleteCookie(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) == 0 {
		t.Fatal("No cookies")
	}
	if err := wd.DeleteCookie(cookies[0].Name); err != nil {
		t.Fatal(err)
	}
	newCookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	if len(newCookies) != len(cookies)-1 {
		t.Fatal("Cookie not deleted")
	}

	for _, c := range newCookies {
		if c.Name == cookies[0].Name {
			t.Fatal("Deleted cookie found")
		}
	}
}

func testLocation(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	button, err := wd.FindElement(ByID, "submit")
	if err != nil {
		t.Fatal(err)
	}

	loc, err := button.Location()
	if err != nil {
		t.Fatal(err)
	}

	if loc.X == 0 || loc.Y == 0 {
		t.Fatalf("Bad location: %v\n", loc)
	}
}

func testLocationInView(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	button, err := wd.FindElement(ByID, "submit")
	if err != nil {
		t.Fatal(err)
	}

	loc, err := button.LocationInView()
	if err != nil {
		t.Fatal(err)
	}

	if loc.X == 0 || loc.Y == 0 {
		t.Fatalf("Bad location: %v\n", loc)
	}
}

func testSize(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	button, err := wd.FindElement(ByID, "submit")
	if err != nil {
		t.Fatal(err)
	}

	size, err := button.Size()
	if err != nil {
		t.Fatal(err)
	}

	if size.Width == 0 || size.Height == 0 {
		t.Fatalf("Bad size: %v\n", size)
	}
}

func testExecuteScript(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	script := "return arguments[0] + arguments[1]"
	args := []interface{}{1, 2}
	reply, err := wd.ExecuteScript(script, args)
	if err != nil {
		t.Fatal(err)
	}

	result, ok := reply.(float64)
	if !ok {
		t.Fatal("Not an int reply")
	}

	if result != 3 {
		t.Fatalf("Bad result %d (expected 3)", int(result))
	}
}

func testExecuteScriptWithNilArgs(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	script := "return document.readyState"
	if _, err := wd.ExecuteScript(script, nil); err != nil {
		t.Fatal(err)
	}
}

func testExecuteScriptOnElement(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	input, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}

	const query = "golang"
	if err := input.SendKeys(query); err != nil {
		t.Fatal(err)
	}

	we, err := wd.FindElement(ByXPATH, "//input[@type='submit']")
	if err != nil {
		t.Fatal(err)
	}

	script := "arguments[0].click()"
	args := []interface{}{we}

	if _, err = wd.ExecuteScript(script, args); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, searchContents) {
		t.Fatalf("Can't find %q on page after searching for %q", searchContents, query)
	}
}

func testScreenshot(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	data, err := wd.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("Empty reply")
	}
}

func testLog(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	url := serverURL + "/log"
	if err := wd.Get(url); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", url, err)
	}
	logs, err := wd.Log(Browser)
	if err != nil {
		t.Fatal(err)
	}

	if len(logs) == 0 {
		t.Fatal("Empty reply")
	}
}

func testIsSelected(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	elem, err := wd.FindElement(ByID, "chuk")
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

	if err := elem.Click(); err != nil {
		t.Fatal("Can't click")
	}

	selected, err = elem.IsSelected()
	if err != nil {
		t.Fatal("Can't get selection")
	}

	if !selected {
		t.Fatal("Not selected")
	}
}

func testIsDisplayed(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	elem, err := wd.FindElement(ByID, "chuk")
	if err != nil {
		t.Fatal("Can't find element")
	}
	displayed, err := elem.IsDisplayed()
	if err != nil {
		t.Fatal("Can't check for displayed")
	}

	if !displayed {
		t.Fatal("Not displayed")
	}
}

func testGetAttributeNotFound(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	elem, err := wd.FindElement(ByID, "chuk")
	if err != nil {
		t.Fatal("Can't find element")
	}

	if _, err = elem.GetAttribute("no-such-attribute"); err == nil {
		t.Fatal("Got non existing attribute")
	}
}

func testMaximizeWindow(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	if err := wd.MaximizeWindow(""); err != nil {
		t.Fatalf("error maximizing window: %s", err)
	}
}

func testResizeWindow(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	if err := wd.ResizeWindow("", 100, 100); err != nil {
		t.Fatalf("error resizing window: %s", err)
	}
}

func testKeyDownUp(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	e, err := wd.FindElement(ByLinkText, "other page")
	if err != nil {
		t.Fatalf("error finding other page link: %v", err)
	}

	if err := wd.KeyDown(ControlKey); err != nil {
		t.Fatalf("error pressing control key down: %v", err)
	}
	if err := e.Click(); err != nil {
		t.Fatalf("error clicking the other page link: %v", err)
	}
	if err := wd.KeyUp(ControlKey); err != nil {
		t.Fatalf("error releasing control key: %v", err)
	}
}

func testCSSProperty(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	e, err := wd.FindElement(ByLinkText, "other page")
	if err != nil {
		t.Fatalf("error finding other page link: %v", err)
	}

	color, err := e.CSSProperty("color")
	if err != nil {
		t.Fatalf(`e.CSSProperty("color") returned error: %v`, err)
	}
	wantColor := "rgba(0, 0, 238, 1)"
	if color != wantColor {
		t.Fatalf(`e.CSSProperty("color") = %q, want %q`, color, wantColor)
	}
}

var homePage = `
<html>
<head>
	<title>Go Selenium Test Suite</title>
</head>
<body>
	The home page. <br />
	<form action="/search">
		<input name="q" /> <input type="submit" id="submit"/> <br />
		<input id="chuk" type="checkbox" /> A checkbox.
	</form>
	Link to the <a href="/other">other page</a>.
</body>
</html>
`

var otherPage = `
<html>
<head>
	<title>Go Selenium Test Suite - Other Page</title>
</head>
<body>
	The other page.
</body>
</html>
`

const searchContents = "The Go Proramming Language"

var searchPage = `
<html>
<head>
	<title>Go Selenium Test Suite - Search Page</title>
</head>
<body>
	You searched for "%s". I'll pretend I've found:
	<p>
	"` + searchContents + `"
	</p>
</body>
</html>
`

var logPage = `
<html>
<head>
	<title>Go Selenium Test Suite - Log Page</title>
	<script>
		console.log("console log");
		throw "exception log";
	</script>
</head>
<body>
	Log test page.
</body>
</html>
`

func handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	page, ok := map[string]string{
		"/":       homePage,
		"/other":  otherPage,
		"/search": searchPage,
		"/log":    logPage,
	}[path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if path == "/search" {
		r.ParseForm()
		page = fmt.Sprintf(page, r.Form["q"][0])
	}
	// Some cookies for the tests
	for i := 0; i < 3; i++ {
		http.SetCookie(w, &http.Cookie{
			Name:  fmt.Sprintf("cookie-%d", i),
			Value: fmt.Sprintf("value-%d", i),
		})
	}
	fmt.Fprintf(w, page)
}
