package selenium

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/google/go-cmp/cmp"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
	"github.com/tebeka/selenium/log"
	"github.com/tebeka/selenium/sauce"
)

var (
	selenium3Path          = flag.String("selenium3_path", "", "The path to the Selenium 3 server JAR. If empty or the file is not present, Firefox tests using Selenium 3 will not be run.")
	firefoxBinarySelenium3 = flag.String("firefox_binary_for_selenium3", "vendor/firefox/firefox", "The name of the Firefox binary for Selenium 3 tests or the path to it. If the name does not contain directory separators, the PATH will be searched.")
	geckoDriverPath        = flag.String("geckodriver_path", "", "The path to the geckodriver binary. If empty or the file is not present, the Geckodriver tests will not be run.")
	javaPath               = flag.String("java_path", "", "The path to the Java runtime binary to invoke. If not specified, 'java' will be used.")

	chromeDriverPath = flag.String("chrome_driver_path", "", "The path to the ChromeDriver binary. If empty or the file is not present, Chrome tests will not be run.")
	chromeBinary     = flag.String("chrome_binary", "vendor/chrome-linux/chrome", "The name of the Chrome binary or the path to it. If name is not an exact path, the PATH will be searched.")

	htmlUnitDriverPath = flag.String("htmlunit_driver_path", "vendor/htmlunit-driver.jar", "The path to the HTMLUnit Driver JAR.")

	useDocker          = flag.Bool("docker", false, "If set, run the tests in a Docker container.")
	runningUnderDocker = flag.Bool("running_under_docker", false, "This is set by the Docker test harness and should not be needed otherwise.")

	startFrameBuffer = flag.Bool("start_frame_buffer", true, "If true, start an Xvfb subprocess and run the browsers in that X server.")

	serverURL string
)

func TestMain(m *testing.M) {
	flag.Parse()
	if err := setDriverPaths(); err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("Exiting early: unable to get the driver paths -- %s", err.Error()))
		os.Exit(1)
	}

	s := httptest.NewServer(http.HandlerFunc(handler))
	serverURL = s.URL
	defer s.Close()
	os.Exit(m.Run())
}

func findBestPath(glob string, binary bool) string {
	matches, err := filepath.Glob(glob)
	if err != nil {
		glog.Warningf("Error globbing %q: %s", glob, err)
		return ""
	}
	if len(matches) == 0 {
		return ""
	}
	// Iterate backwards: newer versions should be sorted to the end.
	sort.Strings(matches)
	for i := len(matches) - 1; i >= 0; i-- {
		path := matches[i]
		fi, err := os.Stat(path)
		if err != nil {
			glog.Warningf("Error statting %q: %s", path, err)
			continue
		}
		if !fi.Mode().IsRegular() {
			continue
		}
		if binary && fi.Mode().Perm()&0111 == 0 {
			continue
		}
		return path
	}
	return ""
}

func setDriverPaths() error {
	if *selenium3Path == "" {
		*selenium3Path = findBestPath("vendor/selenium-server*" /*binary=*/, false)
	}

	if *geckoDriverPath == "" {
		*geckoDriverPath = findBestPath("vendor/geckodriver*" /*binary=*/, true)
	}

	if *chromeDriverPath == "" {
		*chromeDriverPath = findBestPath("vendor/chromedriver*" /*binary=*/, true)
	}

	return nil
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
	sauce               *sauce.Capabilities
	seleniumVersion     semver.Version
	serviceOptions      []ServiceOption
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

	// Chrome-specific tests.
	t.Run("Extension", runTest(testChromeExtension, c))

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the ChromeDriver service: %v", err)
	}
}

func testChromeExtension(t *testing.T, c config) {
	caps := newTestCapabilities(t, c)
	co := caps[chrome.CapabilitiesKey].(chrome.Capabilities)
	const path = "testing/chrome-extension/css_page_red"
	if err := co.AddUnpackedExtension(path); err != nil {
		t.Fatalf("co.AddExtension(%q) returned error: %v", path, err)
	}
	caps[chrome.CapabilitiesKey] = co

	wd, err := NewRemote(caps, c.addr)
	if err != nil {
		t.Fatalf("NewRemote(_, _) returned error: %v", err)
	}
	defer wd.Quit()

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	e, err := wd.FindElement(ByCSSSelector, "body")
	if err != nil {
		t.Fatalf("error finding body: %v", err)
	}

	const property = "background-color"
	color, err := e.CSSProperty(property)
	if err != nil {
		t.Fatalf(`e.CSSProperty(%q) returned error: %v`, property, err)
	}

	const wantColor = "rgba(255, 0, 0, 1)"
	if color != wantColor {
		t.Fatalf("body background has color %q, want %q", color, wantColor)
	}
}

func TestFirefoxSelenium3(t *testing.T) {
	if *useDocker {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(*selenium3Path); err != nil {
		t.Skipf("Skipping Firefox tests using Selenium 3 because Selenium WebDriver JAR not found at path %q", *selenium3Path)
	}
	if _, err := os.Stat(*geckoDriverPath); err != nil {
		t.Skipf("Skipping Firefox tests on Selenium 3 because geckodriver binary %q not found", *geckoDriverPath)
	}

	runFirefoxTests(t, *selenium3Path, config{
		seleniumVersion: semver.MustParse("3.0.0"),
		serviceOptions:  []ServiceOption{GeckoDriver(*geckoDriverPath)},
		path:            *firefoxBinarySelenium3,
	})
}

func TestFirefoxGeckoDriver(t *testing.T) {
	if *useDocker {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(*geckoDriverPath); err != nil {
		t.Skipf("Skipping Firefox tests on Selenium 3 because geckodriver binary %q not found", *geckoDriverPath)
	}

	runFirefoxTests(t, *geckoDriverPath, config{
		path: *firefoxBinarySelenium3,
	})
}

func TestHTMLUnit(t *testing.T) {
	if *useDocker {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(*selenium3Path); err != nil {
		t.Skipf("Skipping HTMLUnit tests because the Selenium WebDriver JAR was not found at path %q", *selenium3Path)
	}
	if _, err := os.Stat(*htmlUnitDriverPath); err != nil {
		t.Skipf("Skipping HTMLUnit tests because the HTMLUnit Driver JAR not found at path %q", *htmlUnitDriverPath)
	}

	if testing.Verbose() {
		SetDebug(true)
	}

	c := config{
		browser:         "htmlunit",
		seleniumVersion: semver.MustParse("3.0.0"),
		serviceOptions:  []ServiceOption{HTMLUnit(*htmlUnitDriverPath)},
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}
	s, err := NewSeleniumService(*selenium3Path, port, c.serviceOptions...)
	if err != nil {
		t.Fatalf("Error starting the WebDriver server with binary %q: %v", *selenium3Path, err)
	}
	c.addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)

	runTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the Selenium service: %v", err)
	}
}

func runFirefoxTests(t *testing.T, webDriverPath string, c config) {
	c.browser = "firefox"

	if s, err := os.Stat(c.path); err != nil || !s.Mode().IsRegular() {
		if path, err := exec.LookPath(c.path); err == nil {
			c.path = path
		} else {
			t.Skipf("Skipping Firefox tests because binary %q not found", c.path)
		}
	}

	if *startFrameBuffer {
		c.serviceOptions = append(c.serviceOptions, StartFrameBuffer())
	}
	if testing.Verbose() {
		SetDebug(true)
		c.serviceOptions = append(c.serviceOptions, Output(os.Stderr))
	}
	if *javaPath != "" {
		c.serviceOptions = append(c.serviceOptions, JavaPath(*javaPath))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}

	var s *Service
	if c.seleniumVersion.Major == 0 {
		s, err = NewGeckoDriverService(webDriverPath, port, c.serviceOptions...)
	} else {
		s, err = NewSeleniumService(webDriverPath, port, c.serviceOptions...)
	}
	if err != nil {
		t.Fatalf("Error starting the WebDriver server with binary %q: %v", webDriverPath, err)
	}

	if c.seleniumVersion.Major == 0 {
		c.addr = fmt.Sprintf("http://127.0.0.1:%d", port)
	} else {
		c.addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
	}

	runFirefoxSubTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the Selenium service: %v", err)
	}
}

func runFirefoxSubTests(t *testing.T, c config) {
	runTests(t, c)

	// Firefox-specific tests.
	t.Run("Preferences", runTest(testFirefoxPreferences, c))
	t.Run("Profile", runTest(testFirefoxProfile, c))
}

func testFirefoxPreferences(t *testing.T, c config) {
	if c.seleniumVersion.Major == 2 {
		t.Skip("This test is known to fail for Selenium 2 and Firefox 47.")
	}
	caps := newTestCapabilities(t, c)
	f, ok := caps[firefox.CapabilitiesKey].(firefox.Capabilities)
	if !ok || f.Prefs == nil {
		f.Prefs = make(map[string]interface{})
	}
	f.Prefs["browser.startup.homepage"] = serverURL
	f.Prefs["browser.startup.page"] = "1"
	caps.AddFirefox(f)

	wd := &remoteWD{
		capabilities: caps,
		urlPrefix:    c.addr,
	}
	defer func() {
		if err := wd.Quit(); err != nil {
			t.Errorf("wd.Quit() returned error: %v", err)
		}
	}()

	if _, err := wd.NewSession(); err != nil {
		t.Fatalf("error in new session - %s", err)
	}

	// TODO(minusnine): use the upcoming Wait API for this.
	var u string
	for i := 0; i < 5; i++ {
		var err error
		u, err = wd.CurrentURL()
		if err != nil {
			t.Fatalf("wd.Current() returned error: %v", err)
		}
		if u == serverURL+"/" {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("wd.Current() = %q, want %q", u, serverURL+"/")
}

func testFirefoxProfile(t *testing.T, c config) {
	if c.seleniumVersion.Major == 2 {
		t.Skip("This test is known to fail for Selenium 2 and Firefox 47.")
	}
	caps := newTestCapabilities(t, c)
	f := caps[firefox.CapabilitiesKey].(firefox.Capabilities)
	const path = "testing/firefox-profile"
	if err := f.SetProfile(path); err != nil {
		t.Fatalf("f.SetProfile(%q) returned error: %v", path, err)
	}
	caps.AddFirefox(f)

	wd := &remoteWD{
		capabilities: caps,
		urlPrefix:    c.addr,
	}
	if _, err := wd.NewSession(); err != nil {
		t.Fatalf("wd.NewSession() returned error: %v", err)
	}
	defer quitRemote(t, wd)

	u, err := wd.CurrentURL()
	if err != nil {
		t.Fatalf("wd.Current() returned error: %v", err)
	}
	const wantURL = "about:config"
	if u != wantURL {
		t.Fatalf("wd.Current() = %q, want %q", u, wantURL)
	}

	// Test that the old Firefox profile location gets migrated for W3C
	// compatibility.
	caps = newW3CCapabilities(map[string]interface{}{"firefox_profile": "base64-encoded Firefox profile goes here"})
	f = caps["alwaysMatch"].(Capabilities)[firefox.CapabilitiesKey].(firefox.Capabilities)
	if f.Profile == "" {
		t.Fatalf("Capability 'firefox_profile' was not migrated to 'moz:firefoxOptions.profile': %+v", caps)
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

func newTestCapabilities(t *testing.T, c config) Capabilities {
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
			},
			W3C: true,
		}
		caps.AddChrome(chrCaps)
	case "firefox":
		f := firefox.Capabilities{}
		if c.path != "" {
			// Selenium 2 uses this option to specify the path to the Firefox binary.
			//caps["firefox_binary"] = c.path
			p, err := filepath.Abs(c.path)
			if err != nil {
				panic(err)
			}
			f.Binary = p
		}
		if testing.Verbose() {
			f.Log = &firefox.Log{
				Level: firefox.Trace,
			}
		}
		caps.AddFirefox(f)
	case "htmlunit":
		caps["javascriptEnabled"] = true
	}

	if c.sauce != nil {
		m, err := c.sauce.ToMap()
		if err != nil {
			t.Fatalf("Error obtaining map for sauce.Capabilities: %s", err)
		}
		for k, v := range m {
			caps[k] = v
		}
		if c.seleniumVersion.Major > 0 {
			caps["seleniumVersion"] = c.seleniumVersion.String()
		}
		caps["name"] = t.Name()
	}

	return caps
}

func newRemote(t *testing.T, c config) WebDriver {
	caps := newTestCapabilities(t, c)
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
	t.Run("DeleteSession", runTest(testDeleteSession, c))
	t.Run("Error", runTest(testError, c))
	t.Run("ExtendedErrorMessage", runTest(testExtendedErrorMessage, c))
	t.Run("Capabilities", runTest(testCapabilities, c))
	t.Run("SetAsyncScriptTimeout", runTest(testSetAsyncScriptTimeout, c))
	t.Run("SetImplicitWaitTimeout", runTest(testSetImplicitWaitTimeout, c))
	t.Run("SetPageLoadTimeout", runTest(testSetPageLoadTimeout, c))
	t.Run("Windows", runTest(testWindows, c))
	t.Run("Get", runTest(testGet, c))
	t.Run("Navigation", runTest(testNavigation, c))
	t.Run("Title", runTest(testTitle, c))
	t.Run("PageSource", runTest(testPageSource, c))
	t.Run("FindElement", runTest(testFindElement, c))
	t.Run("FindElements", runTest(testFindElements, c))
	t.Run("SendKeys", runTest(testSendKeys, c))
	t.Run("Click", runTest(testClick, c))
	t.Run("GetCookies", runTest(testGetCookies, c))
	t.Run("GetCookie", runTest(testGetCookie, c))
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
	t.Run("KeyDownUp", runTest(testKeyDownUp, c))
	t.Run("CSSProperty", runTest(testCSSProperty, c))
	t.Run("Proxy", runTest(testProxy, c))
	t.Run("SwitchFrame", runTest(testSwitchFrame, c))
	t.Run("Wait", runTest(testWait, c))
	t.Run("ActiveElement", runTest(testActiveElement, c))
	t.Run("AcceptAlert", runTest(testAcceptAlert, c))
	t.Run("DismissAlert", runTest(testDismissAlert, c))
}

func testStatus(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	status, err := wd.Status()
	if err != nil {
		t.Fatalf("wd.Status() returned error: %v", err)
	}

	if c.sauce == nil {
		if len(status.OS.Name) == 0 && status.Message == "" {
			t.Fatalf("OS.Name or Message not provided: %+v", status)
		}
	} else if status.Build.Version != "Sauce Labs" {
		t.Fatalf("status.Build.Version = %q, expected 'Sauce Labs'", status.Build.Version)
	}
}

func testNewSession(t *testing.T, c config) {
	// Bypass NewRemote which itself calls NewSession internally.
	wd := &remoteWD{
		capabilities: newTestCapabilities(t, c),
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

	if len(sid) == 0 {
		t.Fatal("Empty session id")
	}

	if wd.id != sid {
		t.Fatal("Session id mismatch")
	}

	if wd.SessionID() != sid {
		t.Fatalf("Got session id mismatch %s != %s", sid, wd.SessionID())
	}

	if c.browser != "htmlunit" && wd.browserVersion.Major == 0 {
		t.Fatalf("wd.browserVersion.Major = %d, expected > 0", wd.browserVersion.Major)
	}
}

func testDeleteSession(t *testing.T, c config) {
	wd := newRemote(t, c)
	if err := DeleteSession(c.addr, wd.SessionID()); err != nil {
		t.Fatalf("DeleteSession(%s, %s) returned error: %v", c.addr, wd.SessionID(), err)
	}
}

func testError(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	_, err := wd.FindElement(ByID, "no-such-element")
	if err == nil {
		t.Fatal("wd.FindElement(ByID, 'no-such-element') did not return an error as expected")
	}

	e, ok := err.(*Error)
	if !ok {
		if c.seleniumVersion.Major > 0 {
			//			t.Skipf("Selenium does not support W3C-style errors.")
		} else {
			t.Fatalf("wd.FindElement(ByID, 'no-such-element') returned an error that is not an *Error: %v", err)
		}
	}
	if want := "no such element"; e.Err != want {
		t.Errorf("wd.FindElement(ByID, 'no-such-element'); err.Err = %q, want %q", e.Err, want)
	}

	var wantCode, wantLegacyCode int
	switch c.browser {
	case "chrome":
		if wd.(*remoteWD).w3cCompatible {
			wantCode = 404
			wantLegacyCode = 0
		} else {
			wantCode = 200
			wantLegacyCode = 7
		}
	case "firefox":
		wantCode = 404
	case "htmlunit":
		wantCode = 500
		wantLegacyCode = 7
	}
	if e.HTTPCode != wantCode {
		t.Errorf("wd.FindElement(ByID, 'no-such-element'); err.HTTPCode = %d, want %d", e.HTTPCode, wantCode)
	}
	if e.LegacyCode != wantLegacyCode {
		t.Errorf("wd.FindElement(ByID, 'no-such-element'); err.LegacyCode = %d, want %d", e.LegacyCode, wantLegacyCode)
	}
}

func testExtendedErrorMessage(t *testing.T, c config) {
	// Bypass NewRemote which itself calls NewSession internally.
	wd := &remoteWD{
		capabilities: newTestCapabilities(t, c),
		urlPrefix:    c.addr,
	}

	var want string
	switch {
	case c.sauce != nil:
		want = "404 Not Found"
	case c.seleniumVersion.Major == 3:
		want = "invalid session id:"
	case c.browser == "firefox" && c.seleniumVersion.Major == 2:
		want = "unknown error:"
	case c.browser == "firefox" && c.seleniumVersion.Major == 0:
		want = "unknown command"
	case c.browser == "chrome":
		want = "invalid session id:"
	default:
		want = "invalid session ID:"
	}
	if err := wd.Close(); err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Fatalf("Got error %q, expected error to start with %q", err, want)
	}
}

// TODO(ekg): does this method work anymore in any browser? It is not part of
// the W3C standard.
func testCapabilities(t *testing.T, c config) {
	if c.browser == "firefox" {
		t.Skip("This method is not supported by Geckodriver.")
	}
	t.Skip("This method crashes Chrome?")
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	caps, err := wd.Capabilities()
	if err != nil {
		t.Fatalf("wd.Capabilities() returned error: %v", err)
	}

	if strings.ToLower(caps["browserName"].(string)) != c.browser {
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

func testWindows(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	firstHandle, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Fatal(err)
	}

	if len(firstHandle) == 0 {
		t.Fatal("Empty handle")
	}

	const linkText = "other page"
	link, err := wd.FindElement(ByLinkText, linkText)
	if err != nil {
		t.Fatalf("wd.FindElement(%q, %q) returned error: %v", ByLinkText, linkText, err)
	}

	switch c.browser {
	case "firefox":
		// Firefox+Geckodriver doesn't handle control characters without appending
		// a terminating null key.
		// https://github.com/mozilla/geckodriver/issues/665
		newWindowModifier := ShiftKey + NullKey
		if err := wd.SendModifier(newWindowModifier /*isDown=*/, true); err != nil {
			t.Fatalf("wd.SendModifer(ShiftKey) returned error: %v", err)
		}
		// Firefox and Geckodriver doesn't handle clicking on an element.
		//
		// https://github.com/mozilla/geckodriver/issues/1007
		if err := link.SendKeys(EnterKey); err != nil {
			t.Fatalf("link.SendKeys(EnterKey) returned error: %v", err)
		}
		if err := wd.SendModifier(newWindowModifier /*isDown=*/, false); err != nil {
			t.Fatalf("wd.SendKeys(ShiftKey) returned error: %v", err)
		}
	case "htmlunit":
		newWindowModifier := ShiftKey
		if err := wd.SendModifier(newWindowModifier /*isDown=*/, true); err != nil {
			t.Fatalf("wd.SendModifer(ShiftKey) returned error: %v", err)
		}
		if err := link.Click(); err != nil {
			t.Fatalf("link.Click() returned error: %v", err)
		}
		if err := wd.SendModifier(newWindowModifier /*isDown=*/, false); err != nil {
			t.Fatalf("wd.SendKeys(ShiftKey) returned error: %v", err)
		}
	case "chrome":
		// Chrome doesn't support handling key events at the browser level.
		// https://crbug.com/chromedriver/30
		otherURL := path.Join(serverURL, "other")
		if _, err := wd.ExecuteScript(fmt.Sprintf("window.open(%q)", otherURL), nil); err != nil {
			t.Fatalf("opening a new window via Javascript returned error: %v", err)
		}
	}

	// Starting a new window can take a while. Try a few times before failing.
	var handles []string
	tries := 5
	for {
		handles, err = wd.WindowHandles()
		if err != nil {
			t.Fatalf("wd.WindowHandles() returned error: %v", err)
		}
		if len(handles) == 2 {
			break
		}
		tries--
		if tries == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	if len(handles) != 2 {
		t.Fatalf("len(wd.WindowHandles()) = %d, expected 2", len(handles))
	}
	var found bool
	var otherHandle string
	for _, h := range handles {
		if h == firstHandle {
			found = true
		} else {
			otherHandle = h
		}
	}
	if !found {
		t.Fatalf("wd.WindowHandles() returned %v, expected to include %q", handles, firstHandle)
	}

	t.Run("SwitchWindow", func(t *testing.T) {
		if err := wd.SwitchWindow(otherHandle); err != nil {
			t.Fatalf("wd.SwitchWindow(firstHandle) returned error: %v", err)
		}
		if _, err := wd.FindElement(ByLinkText, linkText); err == nil {
			t.Fatalf("wd.FindElement(%q, %q) (after opening a new window) returned nil, expected error", ByLinkText, linkText)
		}
		if err := wd.SwitchWindow(firstHandle); err != nil {
			t.Fatalf("wd.SwitchWindow(firstHandle) returned error: %v", err)
		}
		if _, err := wd.FindElement(ByLinkText, linkText); err != nil {
			t.Fatalf("wd.FindElement(%q, %q) (after switching to the original window) returned error: %v", ByLinkText, linkText, err)
		}
	})

	t.Run("MaximizeWindow", func(t *testing.T) {
		if err := wd.MaximizeWindow(otherHandle); err != nil {
			t.Fatalf("error maximizing window: %s", err)
		}
	})

	t.Run("ResizeWindow", func(t *testing.T) {
		if err := wd.ResizeWindow(otherHandle, 100, 100); err != nil {
			t.Fatalf("error resizing window: %s", err)
		}
	})

	t.Run("CloseWindow", func(t *testing.T) {
		if err := wd.CloseWindow(otherHandle); err != nil {
			t.Fatalf("wd.CloseWindow(otherHandle) returned error: %v", err)
		}
		handles, err := wd.WindowHandles()
		if err != nil {
			t.Fatalf("wd.WindowHandles() returned error: %v", err)
		}
		if len(handles) != 1 {
			t.Fatalf("len(wd.WindowHandles()) = %d, expected 1", len(handles))
		}
	})
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

	for _, tc := range []struct {
		by, query string
	}{
		{ByName, "q"},
		{ByCSSSelector, "input[name=q]"},
		{ByXPATH, "/html/body/form/input[1]"},
		{ByLinkText, "тест"},
	} {
		elem, err := wd.FindElement(tc.by, tc.query)
		if err != nil {
			t.Errorf("wd.FindElement(%q, %q) returned error: %v", tc.by, tc.query, err)
			continue
		}

		we, ok := elem.(*remoteWE)
		if !ok {
			t.Errorf("wd.FindElement(%q, %q) = %T, want a *remoteWE", tc.by, tc.query, elem)
			continue
		}

		if len(we.id) == 0 {
			t.Errorf("wd.FindElement(%q, %q) returned an empty element", tc.by, tc.query)
			continue
		}

		if we.parent != wd {
			t.Errorf("wd.FindElement(%q, %q) returned the wrong parent", tc.by, tc.query)
			continue
		}
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
	const searchBoxName = "q"
	input, err := wd.FindElement(ByName, searchBoxName)
	if err != nil {
		t.Fatalf("wd.FindElement(%q, %q) returned error: %v", ByName, searchBoxName, err)
	}
	const query = "golang"
	if err = input.SendKeys(query); err != nil {
		t.Fatalf("input.SendKeys(%q) returned error: %v", query, err)
	}

	const selectTag = "select"
	sel, err := wd.FindElement(ByCSSSelector, selectTag)
	if err != nil {
		t.Fatalf("wd.FindElement(%q, %q) returned error: %v", ByCSSSelector, selectTag, err)
	}
	if err = sel.Click(); err != nil {
		t.Fatalf("input.Click() returned error: %v", err)
	}
	time.Sleep(2 * time.Second)
	option, err := sel.FindElement(ByID, "secondValue")
	if err != nil {
		t.Fatalf("input.FindElement(%q, %q) returned error: %v", ByID, "secondValue", err)
	}
	if err = option.Click(); err != nil {
		t.Fatalf("option.Click() returned error: %v", err)
	}
	if c.browser == "firefox" {
		// Click anywhere else to de-focus the drop-down menu.
		if err = input.Click(); err != nil {
			t.Fatalf("wd.Click() returned error: %v", err)
		}
	}

	const buttonID = "submit"
	button, err := wd.FindElement(ByID, buttonID)
	if err != nil {
		t.Fatalf("wd.FindElement(%q, %q) returned error: %v", ByID, buttonID, err)
	}
	if err := wd.SetPageLoadTimeout(2 * time.Second); err != nil {
		t.Fatalf("wd.SetImplicitWaitTimeout() returned error: %v", err)
	}
	if err = button.Click(); err != nil {
		t.Fatalf("wd.Click() returned error: %v", err)
	}

	// The page load timeout for Firefox and Selenium 3 seems to not work for
	// clicking form buttons. Sleep a bit to reduce impact of a flaky test.
	if c.browser == "firefox" && c.seleniumVersion.Major == 3 {
		time.Sleep(2 * time.Second)
	}
	source, err := wd.PageSource()
	if err != nil {
		t.Fatalf("wd.PageSource() returned error: %v", err)
	}

	if !strings.Contains(source, searchContents) {
		t.Fatalf("Can't find %q on page after searching for %q", searchContents, query)
	}
}

func testGetCookie(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatalf("wd.GetCookies() returned error: %v", err)
	}

	if len(cookies) == 0 {
		t.Fatal("wd.GetCookies() returned no cookies")
	}

	if len(cookies[0].Name) == 0 {
		t.Fatalf("Empty cookie name: %+v", cookies[0])
	}

	got, err := wd.GetCookie(cookies[0].Name)
	if err != nil {
		t.Fatalf("wd.GetCookie(%q) returned error: %v", cookies[0].Name, err)
	}
	if !reflect.DeepEqual(got, cookies[0]) {
		t.Fatalf("wd.GetCookie(%q) = %+v, want %+v", cookies[0].Name, cookies[0], got)
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

func v(s string) semver.Version {
	return semver.MustParse(s)
}

func testAddCookie(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	want := &Cookie{
		Name:   "the nameless cookie",
		Value:  "I have nothing",
		Expiry: math.MaxUint32,
		Domain: "127.0.0.1", // Unlike real browsers, htmlunit requires this to be set.
	}
	if err := wd.AddCookie(want); err != nil {
		t.Fatal(err)
	}

	// These fields are added implicitly by the browser.
	want.Domain = "127.0.0.1"
	want.Path = "/"

	// Firefox and Geckodriver now returns an empty string for the path.
	if c.browser == "firefox" {
		want.Path = ""
		r := wd.(*remoteWD)
		switch {
		// https://github.com/mozilla/geckodriver/issues/463
		case r.browserVersion.LT(v("56.0.0")):
		case c.sauce != nil:
			want.Expiry = 0
		}
	}

	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	var got *Cookie
	for _, cookie := range cookies {
		if cookie.Name == want.Name {
			got = &cookie
			break
		}
	}
	if got == nil {
		t.Fatalf("wd.GetCookies() = %v, missing cookie %q", cookies, want.Name)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("wd.GetCookies() returned diff (-want/+got):\n%s", diff)
	}
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
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
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
	switch {
	case c.browser == "htmlunit":
		t.Skip("Skipping on htmlunit")
	case c.browser == "firefox" && (c.seleniumVersion.Major == 3 || c.seleniumVersion.Major == 0):
		// Log is not supported on Firefox with Selenium 3.
		// https://github.com/w3c/webdriver/issues/406
		// https://github.com/mozilla/geckodriver/issues/284
		t.Skip("The log interface is not supported on Firefox, since it is not yet part of the W3C spec.")
	}
	caps := newTestCapabilities(t, c)
	caps.SetLogLevel(log.Browser, log.All)
	if c.browser == "chrome" {
		caps.SetLogLevel(log.Performance, log.All)
	}

	wd := &remoteWD{
		capabilities: caps,
		urlPrefix:    c.addr,
	}
	if _, err := wd.NewSession(); err != nil {
		t.Fatalf("wd.NewSession() returned error: %v", err)
	}
	defer quitRemote(t, wd)

	url := serverURL + "/log"
	if err := wd.Get(url); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", url, err)
	}
	logs, err := wd.Log(log.Browser)
	if err != nil {
		t.Fatalf("wd.Log(Browser) returned error: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("empty reply from wd.Log(Browser)")
	} else {
		for _, l := range logs {
			if len(l.Level) == 0 || l.Timestamp.Unix() == 0 || len(l.Message) == 0 {
				t.Errorf("wd.Log(Browser) returned malformed message: %+v", l)
			}
		}
	}

	if c.browser == "chrome" {
		logs, err = wd.Log(log.Performance)
		if err != nil {
			t.Fatalf("wd.Log(Performance) returned error: %v", err)
		}
		if len(logs) == 0 {
			t.Fatal("empty reply from wd.Log(Performance)")
		} else {
			for _, l := range logs {
				if len(l.Level) == 0 || l.Timestamp.Unix() == 0 || len(l.Message) == 0 {
					t.Errorf("wd.Log(Browser) returned malformed message: %+v", l)
				}
				// Make sure the timestamp conversion is vaguely correct. In
				// practice, this difference should be in the milliseconds range.
				if time.Now().Sub(l.Timestamp) > time.Hour {
					t.Errorf("Message has timestamp %s > 1 hour ago: %v", l.Timestamp, l)
				}
			}
		}
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
	const id = "chuk"
	elem, err := wd.FindElement(ByID, id)
	if err != nil {
		t.Fatalf("wd.FindElement(ByID, %s) return error %v", id, err)
	}
	displayed, err := elem.IsDisplayed()
	if err != nil {
		t.Fatalf("elem.IsDisplayed() returned error: %v", err)
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

func testActiveElement(t *testing.T, c config) {
	if c.browser == "htmlunit" {
		// TODO(minusnine): figure out why ActiveElement doesn't work in HTMLUnit.
		t.Skip("ActiveElement does not work in HTMLUnit")
	}
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	e, err := wd.ActiveElement()
	if err != nil {
		t.Fatalf("wd.ActiveElement() returned error: %v", err)
	}
	name, err := e.GetAttribute("name")
	if err != nil {
		t.Fatalf("wd.ActiveElement().GetAttribute() returned error: %v", err)
	}
	if name != "q" {
		t.Fatalf("wd.ActiveElement().GetAttribute() returned element with name = %q, expected name = 'q'", name)
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
	if c.browser == "htmlunit" {
		t.Skip("Skipping on htmlunit")
	}
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

	// Later versions of Firefox and HTMLUnit return the "rgb" version.
	wantColors := []string{"rgba(0, 0, 238, 1)", "rgb(0, 0, 238)"}
	for _, wantColor := range wantColors {
		if color == wantColor {
			return
		}
	}
	t.Fatalf(`e.CSSProperty("color") = %q, want one of %q`, color, wantColors)
}

const proxyPageContents = "You are viewing a proxied page"

// addrRewriter rewrites all requsted addresses to the one specified by the
// URL.
type addrRewriter struct{ u *url.URL }

func (a *addrRewriter) Rewrite(ctx context.Context, _ *socks5.Request) (context.Context, *socks5.AddrSpec) {
	port, err := strconv.Atoi(a.u.Port())
	if err != nil {
		panic(err)
	}
	return ctx, &socks5.AddrSpec{
		FQDN: a.u.Hostname(),
		Port: port,
	}
}

func testProxy(t *testing.T, c config) {
	if c.sauce != nil {
		t.Skip("Testing a proxy on Sauce Labs doesn't work.")
	}

	// Create a different web server that should be used if HTTP proxying is enabled.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, proxyPageContents)
	}))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatalf("url.Parse(%q) returned error: %v", s.URL, err)
	}

	t.Run("HTTP", func(t *testing.T) {
		caps := newTestCapabilities(t, c)
		caps.AddProxy(Proxy{
			Type: Manual,
			HTTP: u.Host,
		})
		runTestProxy(t, c, caps)
	})

	t.Run("SOCKS", func(t *testing.T) {
		if c.seleniumVersion.Major == 3 {
			// Selenium 3 fails when converting SOCKSVersion with: "unknown error:
			// java.lang.Long cannot be cast to java.lang.Integer"
			// The fix for this is committed, but not yet in a release.
			//
			// https://github.com/SeleniumHQ/selenium/issues/6917
			t.Skip("Selenium 3 throws an exception with SOCKSVersion type conversion")
		}
		socks, err := socks5.New(&socks5.Config{
			Rewriter: &addrRewriter{u},
		})
		if err != nil {
			t.Fatalf("socks5.New(_) returned error: %v", err)
		}
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("net.Listen(_, _) return error: %v", err)
		}

		// Start serving SOCKS connections, but don't fail the test once the
		// listener is closed at the end of execution.
		done := make(chan struct{})
		go func() {
			err := socks.Serve(l)
			select {
			case <-done:
				return
			default:
			}
			if err != nil {
				t.Fatalf("s.ListenAndServe(_) returned error: %v", err)
			}
		}()
		defer func() {
			close(done)
			l.Close()
		}()

		caps := newTestCapabilities(t, c)
		caps.AddProxy(Proxy{
			Type:         Manual,
			SOCKS:        l.Addr().String(),
			SOCKSVersion: 5,
		})

		runTestProxy(t, c, caps)
	})
}

func runTestProxy(t *testing.T, c config, caps Capabilities) {
	allowProxyForLocalhost(c.browser, caps)

	wd := &remoteWD{
		capabilities: caps,
		urlPrefix:    c.addr,
	}
	defer func() {
		if err := wd.Quit(); err != nil {
			t.Fatalf("wd.Quit() returned error: %v", err)
		}
	}()
	if _, err := wd.NewSession(); err != nil {
		t.Fatalf("wd.NewSession() returned error: %v", err)
	}

	// Request the original server URL.
	if err := wd.Get(serverURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}
	source, err := wd.PageSource()
	if err != nil {
		t.Fatalf("wd.PageSource() returned error: %v", err)
	}

	if !strings.Contains(source, proxyPageContents) {
		if strings.Contains(source, "Go Selenium Test Suite") {
			t.Fatal("Got non-proxied page.")
		}
		t.Fatalf("Got page: %s\n\nExpected: %q", source, proxyPageContents)
	}
}

func allowProxyForLocalhost(browser string, caps Capabilities) {
	switch browser {
	case "firefox":
		// By default, Firefox explicitly does not use a proxy for connection to
		// localhost and 127.0.0.1. Clear this preference to reach our test proxy.
		ff := caps[firefox.CapabilitiesKey].(firefox.Capabilities)
		if ff.Prefs == nil {
			ff.Prefs = make(map[string]interface{})
		}
		ff.Prefs["network.proxy.no_proxies_on"] = ""
		ff.Prefs["network.proxy.allow_hijacking_localhost"] = true
		caps.AddFirefox(ff)

	case "chrome":
		ch := caps[chrome.CapabilitiesKey].(chrome.Capabilities)
		// Allow Chrome to use the specified proxy for localhost, which is
		// needed for the Proxy test. https://crbug.com/899126
		ch.Args = append(ch.Args, "--proxy-bypass-list=<-loopback>")
		caps.AddChrome(ch)
	}
}

func testSwitchFrame(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	if err := wd.Get(serverURL + "/frame"); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", serverURL, err)
	}

	const (
		iframeID      = "iframeID"
		insideFrameID = "chuk"
		outsideDivID  = "outsideOfFrame"
	)

	// Test with the ID of the iframe.
	if err := wd.SwitchFrame(iframeID); err != nil {
		t.Fatalf("wd.SwitchToFrame(%q) returned error: %v", iframeID, err)
	}
	if _, err := wd.FindElement(ByID, insideFrameID); err != nil {
		t.Fatalf("After switching frames using an ID, wd.FindElement(ByID, %q) returned error: %v", insideFrameID, err)
	}
	if _, err := wd.FindElement(ByID, outsideDivID); err == nil {
		t.Fatalf("After switching frames using an ID, wd.FindElement(ByID, %q) returned nil, expected an error", outsideDivID)
	}

	// Test with nil, to return to the top-level context.
	if err := wd.SwitchFrame(nil); err != nil {
		t.Fatalf("wd.SwitchToFrame(nil) returned error: %v", err)
	}
	if _, err := wd.FindElement(ByID, outsideDivID); err != nil {
		t.Fatalf("After switching frames using nil, wd.FindElement(ByID, %q) returned error: %v", outsideDivID, err)
	}

	// Test with a WebElement.
	iframe, err := wd.FindElement(ByID, iframeID)
	if err != nil {
		t.Fatalf("error finding iframe: %v", err)
	}
	if err := wd.SwitchFrame(iframe); err != nil {
		t.Fatalf("wd.SwitchToFrame(nil) returned error: %v", err)
	}
	if _, err := wd.FindElement(ByID, insideFrameID); err != nil {
		t.Fatalf("After switching frames using a WebElement, wd.FindElement(ByID, %q) returned error: %v", insideFrameID, err)
	}
	if _, err := wd.FindElement(ByID, outsideDivID); err == nil {
		t.Fatalf("After switching frames using a WebElement, wd.FindElement(ByID, %q) returned nil, expected an error", outsideDivID)
	}

	// Test with the empty string, to return to the top-level context.
	if err := wd.SwitchFrame(""); err != nil {
		t.Fatalf(`wd.SwitchToFrame("") returned error: %v`, err)
	}
	if _, err := wd.FindElement(ByID, outsideDivID); err != nil {
		t.Fatalf(`After switching frames using "", wd.FindElement(ByID, %q) returned error: %v`, outsideDivID, err)
	}
}

func testWait(t *testing.T, c config) {
	const newTitle = "Title changed."
	titleChangeCondition := func(wd WebDriver) (bool, error) {
		title, err := wd.Title()
		if err != nil {
			return false, err
		}

		return title == newTitle, nil
	}

	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	titleURL := serverURL + "/title"

	if err := wd.Get(titleURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", titleURL, err)
	}

	// Testing when the title should change.
	if err := wd.Wait(titleChangeCondition); err != nil {
		t.Fatalf("wd.Wait(titleChangeCondition) returned error: %v", err)
	}

	// Reloading the page.
	if err := wd.Get(titleURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", titleURL, err)
	}

	// Testing when the Wait() should error the timeout..
	if err := wd.WaitWithTimeout(titleChangeCondition, 500*time.Millisecond); err == nil {
		t.Fatalf("wd.Wait(titleChangeCondition) should returned error, but it didn't.")
	}
}

func testAcceptAlert(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	alertPageURL := serverURL + "/alert"

	if err := wd.Get(alertPageURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", alertPageURL, err)
	}

	alertText, err := wd.AlertText()
	if err != nil {
		t.Fatalf("wd.AlertText() returned error: %v", err)
	}

	if alertText != "Hello world" {
		t.Fatalf("Expected 'Hello world' but got '%s'", alertText)
	}

	if err := wd.AcceptAlert(); err != nil {
		t.Fatalf("wd.AcceptAlert() returned error: %v", err)
	}
}

func testDismissAlert(t *testing.T, c config) {
	wd := newRemote(t, c)
	defer quitRemote(t, wd)

	alertPageURL := serverURL + "/alert"

	if err := wd.Get(alertPageURL); err != nil {
		t.Fatalf("wd.Get(%q) returned error: %v", alertPageURL, err)
	}

	if err := wd.DismissAlert(); err != nil {
		t.Fatalf("wd.DismissAlert() returned error: %v", err)
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
		<input name="q" autofocus />
		<input type="submit" id="submit" /> <br />
		<input id="chuk" type="checkbox" /> A checkbox.
		<select name="s">
			<option value="first_value">First Value</option>
			<option id="secondValue" value="second_value">Second Value</option>
		</select>
	</form>
	Link to the <a href="/other">other page</a>.

	<a href="/log">тест</a>
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
	Select value is: %s
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

var framePage = `
<html>
<head>
	<title>Go Selenium Test Suite - Frame Page</title>
</head>
<body>
	This page contains a frame.

	<iframe id="iframeID" name="iframeName" src="/"></iframe>
	<div id="outsideOfFrame"></div>
</body>
</html>
`

var titleChangePage = `
<html>
<head>
	<title>Go Selenium Test Suite - Title Change Page</title>
</head>
<body>
	This page will change a title after 1 second.

	<script>
		setTimeout(function() { document.title = 'Title changed.' }, 1000);
	</script>
</body>
</html>
`

var alertPage = `
<html>
<head>
	<title>Go Selenium Test Suite - Alert Appear Page</title>
</head>
<body>
	An alert will popup.

	<script>
		alert('Hello world');
	</script>
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
		"/frame":  framePage,
		"/title":  titleChangePage,
		"/alert":  alertPage,
	}[path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if path == "/search" {
		r.ParseForm()
		page = fmt.Sprintf(page, r.Form["q"][0], r.Form["s"][0])
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
