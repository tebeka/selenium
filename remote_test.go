package selenium_test

import (
	"flag"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/internal/seleniumtest"
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

	startFrameBuffer = flag.Bool("start_frame_buffer", false, "If true, start an Xvfb subprocess and run the browsers in that X server.")
	headless         = flag.Bool("headless", true, "If true, run Chrome and Firefox in headless mode, not requiring a frame buffer.")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if err := setDriverPaths(); err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("Exiting early: unable to get the driver paths -- %s", err.Error()))
		os.Exit(1)
	}
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

	t.Run("Chromedriver", func(t *testing.T) {
		runChromeTests(t, seleniumtest.Config{
			Path: *chromeBinary,
		})
	})

	t.Run("Selenium3", func(t *testing.T) {
		runChromeTests(t, seleniumtest.Config{
			Path:            *chromeBinary,
			SeleniumVersion: semver.MustParse("3.0.0"),
		})
	})
}

func runChromeTests(t *testing.T, c seleniumtest.Config) {
	c.Browser = "chrome"
	c.Headless = *headless

	var opts []selenium.ServiceOption
	if *startFrameBuffer {
		opts = append(opts, selenium.StartFrameBuffer())
	}
	if testing.Verbose() {
		selenium.SetDebug(true)
		opts = append(opts, selenium.Output(os.Stderr))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}
	c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)

	var s *selenium.Service
	if c.SeleniumVersion.Major == 3 {
		c.ServiceOptions = append(c.ServiceOptions, selenium.ChromeDriver(*chromeDriverPath))
		s, err = selenium.NewSeleniumService(*selenium3Path, port, c.ServiceOptions...)
	} else {
		s, err = selenium.NewChromeDriverService(*chromeDriverPath, port, c.ServiceOptions...)
	}
	if err != nil {
		t.Fatalf("Error starting the server: %v", err)
	}

	hs := httptest.NewServer(seleniumtest.Handler)
	defer hs.Close()
	c.ServerURL = hs.URL

	seleniumtest.RunCommonTests(t, c)
	seleniumtest.RunChromeTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the ChromeDriver service: %v", err)
	}
}

func TestFirefox(t *testing.T) {
	if *useDocker {
		t.Skip("Skipping tests because they will be run under a Docker container")
	}
	if _, err := os.Stat(*geckoDriverPath); err != nil {
		t.Skipf("Skipping Firefox tests on Selenium 3 because geckodriver binary %q not found", *geckoDriverPath)
	}

	if s, err := os.Stat(*firefoxBinarySelenium3); err != nil || !s.Mode().IsRegular() {
		if p, err := exec.LookPath(*firefoxBinarySelenium3); err == nil {
			*firefoxBinarySelenium3 = p
		} else {
			t.Skipf("Skipping Firefox tests because binary %q not found", *firefoxBinarySelenium3)
		}
	}

	t.Run("Selenium3", func(t *testing.T) {
		runFirefoxTests(t, *selenium3Path, seleniumtest.Config{
			SeleniumVersion: semver.MustParse("3.0.0"),
			ServiceOptions:  []selenium.ServiceOption{selenium.GeckoDriver(*geckoDriverPath)},
			Path:            *firefoxBinarySelenium3,
		})
	})
	t.Run("Geckodriver", func(t *testing.T) {
		runFirefoxTests(t, *geckoDriverPath, seleniumtest.Config{
			Path: *firefoxBinarySelenium3,
		})
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
		selenium.SetDebug(true)
	}

	c := seleniumtest.Config{
		Browser:         "htmlunit",
		SeleniumVersion: semver.MustParse("3.0.0"),
		ServiceOptions:  []selenium.ServiceOption{selenium.HTMLUnit(*htmlUnitDriverPath)},
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}
	s, err := selenium.NewSeleniumService(*selenium3Path, port, c.ServiceOptions...)
	if err != nil {
		t.Fatalf("Error starting the WebDriver server with binary %q: %v", *selenium3Path, err)
	}
	c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)

	hs := httptest.NewServer(seleniumtest.Handler)
	defer hs.Close()
	c.ServerURL = hs.URL

	seleniumtest.RunCommonTests(t, c)

	if err := s.Stop(); err != nil {
		t.Fatalf("Error stopping the Selenium service: %v", err)
	}
}

func runFirefoxTests(t *testing.T, webDriverPath string, c seleniumtest.Config) {
	c.Browser = "firefox"

	if *startFrameBuffer {
		c.ServiceOptions = append(c.ServiceOptions, selenium.StartFrameBuffer())
	}
	if testing.Verbose() {
		selenium.SetDebug(true)
		c.ServiceOptions = append(c.ServiceOptions, selenium.Output(os.Stderr))
	}
	if *javaPath != "" {
		c.ServiceOptions = append(c.ServiceOptions, selenium.JavaPath(*javaPath))
	}

	port, err := pickUnusedPort()
	if err != nil {
		t.Fatalf("pickUnusedPort() returned error: %v", err)
	}

	var s *selenium.Service
	if c.SeleniumVersion.Major == 0 {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d", port)
		s, err = selenium.NewGeckoDriverService(webDriverPath, port, c.ServiceOptions...)
	} else {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
		if _, err := os.Stat(*selenium3Path); err != nil {
			t.Skipf("Skipping Firefox tests using Selenium 3 because Selenium WebDriver JAR not found at path %q", *selenium3Path)
		}

		s, err = selenium.NewSeleniumService(webDriverPath, port, c.ServiceOptions...)
	}
	if err != nil {
		t.Fatalf("Error starting the WebDriver server with binary %q: %v", webDriverPath, err)
	}

	hs := httptest.NewServer(seleniumtest.Handler)
	defer hs.Close()
	c.ServerURL = hs.URL

	if c.SeleniumVersion.Major == 0 {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d", port)
	} else {
		c.Addr = fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
	}

	c.Headless = *headless

	seleniumtest.RunCommonTests(t, c)
	seleniumtest.RunFirefoxTests(t, c)

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

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() returned error: %v", err)
	}

	// TODO(minusnine): pass through relevant flags to docker-test.sh to be
	// passed to go test.
	cmd := exec.Command("docker", "run", fmt.Sprintf("--volume=%s:/code", cwd), "--workdir=/code/", "go-selenium", "testing/docker-test.sh")
	if testing.Verbose() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker run failed: %v", err)
	}
}
